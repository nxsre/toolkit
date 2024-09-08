package main

import (
	"context"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"github.com/jessevdk/go-flags"
	"github.com/nxsre/toolkit/pkg/clusterpb"

	//"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/lib/v4/cache"
	ristretto_store "github.com/eko/gocache/store/ristretto/v4"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/log"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/memberlist"
	jsoniter "github.com/json-iterator/go"
	"github.com/nxsre/toolkit/pkg/cluster"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rosedblabs/wal"
	"io"
	"net"
	"os"
	"path/filepath"
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

	// 本地缓存
	cacheCtx := context.Background()
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // 最大可用 cost (1GB).  对应 set 的 cost 值，每个 key 的 cost 累计
		BufferItems: 64,      // number of keys per Get buffer.
		Metrics:     true,
	})
	if err != nil {
		panic(err)
	}

	wallog, err := wal.Open(wal.Options{
		DirPath:        filepath.Join(fmt.Sprint(opts.Port)),
		SegmentSize:    wal.KB,
		SegmentFileExt: ".SEG",
		Sync:           false,
		BytesPerSync:   0,
	})

	// 模拟本地存储
	ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)

	cacheManager := cache.New[string](ristrettoStore)
	var process = func(data []byte) {
		// 写日志
		wallog.Write(data)

		u := update{}
		_ = jsoniter.Unmarshal(data, &u)
		switch u.Action {
		case "add":
			err := cacheManager.Set(cacheCtx, u.Data["key"], u.Data["val"])
			if err != nil {
				fmt.Println("add err:", err)
				return
			}
		case "del":
			err := cacheManager.Delete(cacheCtx, u.Data["key"])
			if err != nil {
				fmt.Println("del err:", err)
				return
			}
		}
	}

	// 程序启动先从 wallog 加载数据到内存
	reader := wallog.NewReader()
	for {
		val, pos, err := reader.Next()
		if err == io.EOF {
			break
		}
		process(val)
		fmt.Println("pos::", pos) // get position of the data for next read
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
		val, err := cacheManager.Get(cacheCtx, ctx.Query("key"))
		if err != nil {
			fmt.Println("get", val, err)
		}
		ctx.JSON(200, val)
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
