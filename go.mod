module github.com/nxsre/toolkit

go 1.22.5

require (
	github.com/Microsoft/hcsshim v0.12.7
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/avast/retry-go/v4 v4.6.0
	github.com/breml/rootcerts v0.2.17
	github.com/c3sr/cpu v1.0.0
	github.com/cilium/ebpf v0.16.0
	github.com/containerd/accelerated-container-image v1.1.4
	github.com/containerd/cgroups/v3 v3.0.3
	github.com/containerd/console v1.0.4
	github.com/containerd/containerd v1.7.22
	github.com/containerd/containerd/api v1.8.0-rc.4
	github.com/containerd/continuity v0.4.3
	github.com/containerd/errdefs v0.1.0
	github.com/containerd/go-cni v1.1.10
	github.com/containerd/imgcrypt v1.1.11
	github.com/containerd/log v0.1.0
	github.com/containerd/nerdctl v1.7.7
	github.com/containerd/nydus-snapshotter v0.14.0
	github.com/containerd/platforms v0.2.1
	github.com/containerd/stargz-snapshotter v0.15.1
	github.com/containerd/stargz-snapshotter/estargz v0.15.2-0.20240709063920-1dac5ef89319
	github.com/containerd/typeurl/v2 v2.2.0
	github.com/cyphar/filepath-securejoin v0.3.4
	github.com/dgraph-io/ristretto v0.1.2-0.20240116140435-c67e07994f91
	github.com/distribution/reference v0.6.0
	github.com/docker/cli v27.3.1+incompatible
	github.com/docker/docker v27.3.1+incompatible
	github.com/docker/go-units v0.5.0
	github.com/dustin/go-humanize v1.0.1
	github.com/eko/gocache/lib/v4 v4.1.6
	github.com/eko/gocache/store/ristretto/v4 v4.2.2
	github.com/emersion/go-sasl v0.0.0-20231106173351-e73c9f7bad43
	github.com/emersion/go-smtp v0.21.3
	github.com/gabriel-vasile/mimetype v1.4.5
	github.com/gin-gonic/gin v1.10.0
	github.com/go-acme/lego/v4 v4.19.0
	github.com/go-co-op/gocron/v2 v2.11.0
	github.com/go-hermes/hermes/v2 v2.3.0
	github.com/go-kit/log v0.2.1
	github.com/go-resty/resty/v2 v2.13.1
	github.com/gogo/protobuf v1.3.2
	github.com/golang/glog v1.2.2
	github.com/hanwen/go-fuse/v2 v2.5.1
	github.com/hashicorp/go-retryablehttp v0.7.7
	github.com/hashicorp/go-sockaddr v1.0.2
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/hashicorp/memberlist v0.5.0
	github.com/imroc/req/v3 v3.46.0
	github.com/jedib0t/go-pretty/v6 v6.5.9
	github.com/jessevdk/go-flags v1.6.1
	github.com/jochenvg/go-udev v0.0.0-20240801134859-b65ed646224b
	github.com/json-iterator/go v1.1.12
	github.com/klauspost/compress v1.17.11
	github.com/kr/pretty v0.3.1
	github.com/lrita/numa v1.0.2
	github.com/melbahja/goph v1.4.0
	github.com/moby/spdystream v0.4.0
	github.com/moby/sys/signal v0.7.1
	github.com/moby/sys/userns v0.1.0
	github.com/oklog/ulid v1.3.1
	github.com/ollama/ollama v0.3.9
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0
	github.com/opencontainers/runtime-spec v1.2.0
	github.com/panjf2000/ants/v2 v2.10.0
	github.com/pilebones/go-udev v0.9.0
	github.com/prometheus/client_golang v1.20.5
	github.com/prometheus/client_model v0.6.1
	github.com/prometheus/common v0.55.0
	github.com/prometheus/exporter-toolkit v0.10.0
	github.com/rosedblabs/wal v1.3.8
	github.com/shirou/gopsutil/v4 v4.24.8
	github.com/sirupsen/logrus v1.9.3
	github.com/smantriplw/fasthttp-reverse-proxy/v2 v2.0.0-20240505083843-68b83898d9e1
	github.com/stretchr/testify v1.9.0
	github.com/tmc/langchaingo v0.1.12
	github.com/u-root/u-root v0.14.0
	github.com/valyala/fasthttp v1.55.0
	github.com/veqryn/slog-context v0.7.0
	github.com/veqryn/slog-context/otel v0.7.0
	github.com/vishvananda/netlink v1.3.0
	github.com/vishvananda/netns v0.0.4
	github.com/wneessen/go-mail v0.4.3
	github.com/xplorfin/fasthttp2curl v0.28.0
	github.com/xuri/excelize/v2 v2.8.1
	go.etcd.io/bbolt v1.3.11
	go.opentelemetry.io/otel v1.31.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.29.0
	go.opentelemetry.io/otel/sdk v1.31.0
	go.opentelemetry.io/otel/trace v1.31.0
	golang.org/x/crypto v0.28.0
	golang.org/x/net v0.30.0
	golang.org/x/sys v0.26.0
	golang.org/x/term v0.25.0
	golang.org/x/text v0.19.0
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/sqlite v1.5.6
	gorm.io/gorm v1.25.12
	gorm.io/plugin/prometheus v0.1.0
	k8s.io/api v0.31.1
	k8s.io/apimachinery v0.31.1
	k8s.io/cli-runtime v0.28.8
	k8s.io/client-go v0.31.1
	k8s.io/klog/v2 v2.130.1
	k8s.io/kubectl v0.0.0
	k8s.io/kubernetes v1.28.8
	k8s.io/utils v0.0.0-20240711033017-18e509b52bc8
)

require (
	cloud.google.com/go v0.115.1 // indirect
	cloud.google.com/go/auth v0.9.3 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.4 // indirect
	cloud.google.com/go/compute/metadata v0.5.1 // indirect
	cloud.google.com/go/firestore v1.16.0 // indirect
	cloud.google.com/go/longrunning v0.6.0 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20240806141605-e8a1dd7889d6 // indirect
	github.com/AdamKorcz/go-118-fuzz-build v0.0.0-20231105174938-2b5cbb29f3e2 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/GeertJohan/go-sourcepath v0.0.0-20150925135350-83e8b8723a9b // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/semver/v3 v3.3.0 // indirect
	github.com/Masterminds/sprig v2.16.0+incompatible // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/PuerkitoBio/goquery v1.9.1 // indirect
	github.com/aliyun/alibaba-cloud-sdk-go v1.63.15 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/aokoli/goutils v1.0.1 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/awslabs/soci-snapshotter v0.4.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/c3sr/config v1.0.1 // indirect
	github.com/c3sr/logger v1.0.1 // indirect
	github.com/c3sr/utils v1.0.0 // indirect
	github.com/c3sr/vipertags v1.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cheekybits/genny v1.0.0 // indirect
	github.com/cloudflare/circl v1.4.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/compose-spec/compose-go v1.20.2 // indirect
	github.com/containerd/fifo v1.1.0 // indirect
	github.com/containerd/go-runc v1.1.0 // indirect
	github.com/containerd/stargz-snapshotter/ipfs v0.15.1 // indirect
	github.com/containerd/ttrpc v1.2.6-0.20240827082320-b5cd6e4b3287 // indirect
	github.com/containerd/typeurl v1.0.3-0.20220422153119-7f6e6d160d67 // indirect
	github.com/containernetworking/cni v1.2.3 // indirect
	github.com/containernetworking/plugins v1.5.1 // indirect
	github.com/containers/ocicrypt v1.2.0 // indirect
	github.com/coreos/go-iptables v0.8.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/djherbis/times v1.6.0 // indirect
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/docker/docker-credential-helpers v0.8.2 // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/fahedouch/go-logrotate v0.2.1 // indirect
	github.com/fasthttp/websocket v1.5.8 // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fluent/fluent-logger-golang v1.9.0 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/fvbommel/sortorder v1.1.0 // indirect
	github.com/gin-contrib/pprof v1.5.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.20.2 // indirect
	github.com/go-openapi/jsonreference v0.20.4 // indirect
	github.com/go-openapi/swag v0.22.9 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.20.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.7.0-rc.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/cadvisor v0.47.3 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20240910150728-a0b0bb1d4134 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/hashicorp/consul/api v1.25.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.6.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/intel-go/cpuid v0.0.0-20200819041909-2aa72927c3e2 // indirect
	github.com/ipfs/go-cid v0.4.1 // indirect
	github.com/jaytaylor/html2text v0.0.0-20230321000545-74c2419ad056 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jkeiser/iter v0.0.0-20200628201005-c8aa0ae784d1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/joho/godotenv v1.3.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/k0kubun/pp/v3 v3.2.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/miekg/dns v1.1.62 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/sys/mount v0.3.4 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/symlink v0.3.0 // indirect
	github.com/moby/sys/user v0.3.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr v0.13.0 // indirect
	github.com/multiformats/go-multibase v0.2.0 // indirect
	github.com/multiformats/go-multihash v0.2.3 // indirect
	github.com/multiformats/go-varint v0.0.7 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/nats-io/nats.go v1.31.0 // indirect
	github.com/nats-io/nkeys v0.4.6 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/onsi/ginkgo/v2 v2.20.2 // indirect
	github.com/opencontainers/selinux v1.11.1 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/philhofer/fwd v1.1.3-0.20240612014219-fbbf4953d986 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.6 // indirect
	github.com/pkoukk/tiktoken-go v0.1.6 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/quic-go/quic-go v0.47.0 // indirect
	github.com/refraction-networking/utls v1.6.7 // indirect
	github.com/richardlehane/mscfb v1.0.4 // indirect
	github.com/richardlehane/msoleps v1.0.3 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	github.com/rootless-containers/bypass4netns v0.4.1 // indirect
	github.com/rootless-containers/rootlesskit v1.1.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/crypt v0.17.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/savsgio/gotils v0.0.0-20240303185622-093b76447511 // indirect
	github.com/sean-/seed v0.0.0-20170313163322-e2103e2c3529 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/cobra v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.18.2 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stefanberger/go-pkcs11uri v0.0.0-20230803200340-78284954bff6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.0.1002 // indirect
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod v1.0.1002 // indirect
	github.com/tidwall/gjson v1.17.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tinylib/msgp v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/unknwon/com v1.0.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/vanng822/css v1.0.1 // indirect
	github.com/vanng822/go-premailer v1.21.0 // indirect
	github.com/vbatts/tar-split v0.11.5 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/xuri/efp v0.0.0-20231025114914-d1ff6096ae53 // indirect
	github.com/xuri/nfp v0.0.0-20230919160717-d98342af3f05 // indirect
	github.com/yuchanns/srslog v1.1.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.etcd.io/etcd/api/v3 v3.5.13 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.5.13 // indirect
	go.etcd.io/etcd/client/v2 v2.305.10 // indirect
	go.etcd.io/etcd/client/v3 v3.5.13 // indirect
	go.mozilla.org/pkcs7 v0.0.0-20210826202110-33d05740a352 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.56.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.56.0 // indirect
	go.opentelemetry.io/otel/metric v1.31.0 // indirect
	go.starlark.net v0.0.0-20240411212711-9b43f0afd521 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/mock v0.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/arch v0.8.0 // indirect
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/image v0.18.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/oauth2 v0.23.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/time v0.6.0 // indirect
	golang.org/x/tools v0.25.0 // indirect
	google.golang.org/api v0.197.0 // indirect
	google.golang.org/genproto v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241007155032-5fefd90f89a9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241007155032-5fefd90f89a9 // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/component-base v0.31.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340 // indirect
	lukechampine.com/blake3 v1.3.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/api v0.17.2 // indirect
	sigs.k8s.io/kustomize/kyaml v0.17.1 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace k8s.io/api => k8s.io/api v0.28.8

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.28.8

replace k8s.io/apimachinery => k8s.io/apimachinery v0.28.8

replace k8s.io/apiserver => k8s.io/apiserver v0.28.8

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.28.8

replace k8s.io/client-go => k8s.io/client-go v0.28.8

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.28.8

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.28.8

replace k8s.io/code-generator => k8s.io/code-generator v0.28.8

replace k8s.io/component-base => k8s.io/component-base v0.28.8

replace k8s.io/component-helpers => k8s.io/component-helpers v0.28.8

replace k8s.io/controller-manager => k8s.io/controller-manager v0.28.8

replace k8s.io/cri-api => k8s.io/cri-api v0.28.8

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.28.8

replace k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.28.8

replace k8s.io/endpointslice => k8s.io/endpointslice v0.28.8

replace k8s.io/kms => k8s.io/kms v0.28.8

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.28.8

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.28.8

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.28.8

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.28.8

replace k8s.io/kubectl => k8s.io/kubectl v0.28.8

replace k8s.io/kubelet => k8s.io/kubelet v0.28.8

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.28.8

replace k8s.io/metrics => k8s.io/metrics v0.28.8

replace k8s.io/mount-utils => k8s.io/mount-utils v0.28.8

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.28.8

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.28.8

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.28.8

replace k8s.io/sample-controller => k8s.io/sample-controller v0.28.8

replace github.com/apache/rocketmq-client-go/v2 => /opt/workspaces/tuxedo_rmq_client_go/

replace github.com/matt-e/go-adb => /opt/workspaces/go-adb

replace github.com/shogo82148/androidbinary => /opt/workspaces/androidbinary

replace github.com/zclwy/apkparser => /opt/workspaces/cph/apkparser

replace github.com/emersion/go-smtp => /opt/workspaces/go-smtp

replace (
	github.com/emersion/go-sasl => /opt/workspaces/go-sasl
	github.com/emersion/go-sasl v0.0.0-20231106173351-e73c9f7bad43 => /opt/workspaces/go-sasl
)
