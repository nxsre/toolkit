package main

import (
	"context"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/nxsre/cluster/pkg/clusterpb"
	bolt "go.etcd.io/bbolt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/memberlist"
	jsoniter "github.com/json-iterator/go"
	"github.com/nxsre/cluster/pkg/cluster"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"os"
	"time"
)

type update struct {
	Action string
	Data   map[string]string
}

func main() {
	var opts struct {
		Port  uint     `long:"port" short:"p" description:"Port to listen on" `
		Peers []string `long:"peers" description:"peers" `
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
	}

	db, err := bolt.Open(filepath.Join(fmt.Sprint("_", opts.Port)), 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 创建 bucket
	{
		// Start a writable transaction.
		tx, err := db.Begin(true)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer tx.Rollback()

		// Use the transaction...
		_, err = tx.CreateBucketIfNotExists([]byte("MyBucket"))
		if err != nil {
			fmt.Println(err)
			return
		}

		// Commit the transaction and check for error.
		if err := tx.Commit(); err != nil {
			fmt.Println(err)
			return
		}
	}

	var process = func(data []byte) {

		u := update{}
		_ = jsoniter.Unmarshal(data, &u)
		switch u.Action {
		case "add":
			err := db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("MyBucket"))
				err := b.Put([]byte(u.Data["key"]), []byte(u.Data["val"]))
				return err
			})
			if err != nil {
				fmt.Println("add", err)
				return
			}
		case "del":
			err := db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("MyBucket"))
				err := b.Delete([]byte(u.Data["key"]))
				return err
			})
			if err != nil {
				fmt.Println("del", err)
				return
			}
		}
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	// 注册 prometheus 并开启 exporter
	reg := prometheus.NewRegistry()
	prometheus.MustRegister(reg)
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 集群消息处理
	var notifyMsg = func(b []byte) error {
		m := clusterpb.Part{}
		err := proto.Unmarshal(b, &m)
		if err != nil {
			return err
		}
		fmt.Println("收到消息", m.Key)

		process(m.Data)
		return nil
	}

	// 创建集群
	p, err := cluster.Create(
		logger,
		reg,
		fmt.Sprintf("127.0.0.1:%d", opts.Port),
		"",
		opts.Peers,
		true,
		cluster.DefaultPushPullInterval,
		cluster.DefaultGossipInterval,
		cluster.DefaultTCPTimeout,
		cluster.DefaultProbeTimeout,
		cluster.DefaultProbeInterval,
		nil,
		false,
		"xxx",
		notifyMsg,
	)

	if err != nil {
		fmt.Println("---", err)
		//os.Exit(1)
	}

	err = p.Join(
		cluster.DefaultReconnectInterval,
		cluster.DefaultReconnectTimeout,
	)

	if err != nil {
		fmt.Println("===", err)
		//return
	}
	// Settle 等待一段时间，用于等待成员加入集群上线
	p.Settle(context.Background(), 0*time.Second)

	err = p.WaitReady(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(p.Status(), p.Info(), p.Peers()[0].Name(), err)

	retransmit := len(p.Peers()) / 2
	if retransmit < 3 {
		retransmit = 8
	}

	c := newChannel(
		// send
		// 小于 Oversize 调用 SendBestEffort 使用 udp 发消息
		func(b []byte) {
			// 发广播消息
			p.SendBestEffort(b)
		},
		// peers
		// 超过 Oversize 调用 peers 获取接收消息的成员列表
		func() []*memberlist.Node {
			fmt.Println("get peers")
			nodes := []*memberlist.Node{}
			for _, peer := range p.Peers() {
				// 不给自己发
				if peer.Name() != p.Self().Name {
					addr, _ := net.ResolveUDPAddr("udp", peer.Address())
					nodes = append(nodes, &memberlist.Node{Addr: net.ParseIP(addr.IP.String()), Port: uint16(addr.Port), Name: peer.Name()})
				}
			}
			return nodes
		},
		// sendOversize
		// 超过 Oversize 调用 SendReliable 使用 tcp 发消息
		func(n *memberlist.Node, b []byte) error {
			fmt.Println("call sendOversize to ", n.String(), n.Address(), len(b))
			p.SendReliable(n, b)
			return nil
		},
		reg,
	)

	r.GET("/add", func(ctx *gin.Context) {
		bs, _ := jsoniter.Marshal(update{Action: "add", Data: map[string]string{
			"key": ctx.Query("key"),
			"val": ctx.Query("val"),
		}})
		// 当前节点处理消息
		process(bs)

		// 广播到其他节点
		c.Broadcast(bs)
		ctx.JSON(200, gin.H{})
	})

	r.GET("/del", func(ctx *gin.Context) {
		bs, _ := jsoniter.Marshal(update{Action: "del", Data: map[string]string{
			"key": ctx.Query("key"),
		}})
		// 当前节点处理消息
		process(bs)

		// 广播到其他节点
		c.Broadcast(bs)
		ctx.JSON(200, gin.H{})
	})

	r.GET("/get", func(ctx *gin.Context) {
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("MyBucket"))
			v := b.Get([]byte(ctx.Query("key")))
			fmt.Printf("The answer is: %s\n", v)
			ctx.JSON(200, string(v))
			return nil
		})
	})
	r.GET("/backup", func(ctx *gin.Context) {
		// 备份数据库
		err := db.View(func(tx *bolt.Tx) error {
			ctx.Header("Content-Type", "application/octet-stream")
			ctx.Header("Content-Disposition", `attachment; filename="my.db"`)
			ctx.Header("Content-Length", strconv.Itoa(int(tx.Size())))
			_, err := tx.WriteTo(ctx.Writer)
			return err
		})
		if err != nil {
			http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
		}
	})

	if err := r.Run(fmt.Sprintf(":%d", opts.Port+10000)); err != nil {
		panic(err)
	}
}

func newChannel(
	send func([]byte),
	peers func() []*memberlist.Node,
	sendOversize func(*memberlist.Node, []byte) error,
	registerer prometheus.Registerer,
) *cluster.Channel {
	return cluster.NewChannel(
		"test",
		send,
		peers,
		sendOversize,
		log.NewJSONLogger(os.Stdout),
		make(chan struct{}),
		registerer,
	)
}
