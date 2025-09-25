openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 3650 -nodes -subj "/C=XX/ST=StateName/L=CityName/O=CompanyName/OU=CompanySectionName/CN=CommonNameOrHostname"


NAMESPACE=openshift-sriov-network-operator go run ./cmd/webhook start --tls-cert-file=cert.pem --tls-private-key-file=key.pem --port=6443 --alsologtostderr=true 
2025-09-23T10:40:33.006020594+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     Run sriov-network-operator-webhook
2025-09-23T10:40:33.511867252+02:00     ERROR   sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     failed to retrieve supported NICs       {"error": "configmaps \"supported-nic-ids\" not found"}
2025-09-23T10:40:33.512768211+02:00     INFO    sriov-network-operator-webhook  runtime/asm_amd64.s:1700        start server
2025-09-23T10:40:43.404990855+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     watcher event   {"event": "WRITE         \"key.pem\""}
2025-09-23T10:40:43.40506446+02:00      INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     modified file   {"name": "key.pem"}
2025-09-23T10:40:43.405299907+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     watcher event   {"event": "WRITE         \"key.pem\""}
2025-09-23T10:40:43.405377172+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     modified file   {"name": "key.pem"}
2025-09-23T10:40:43.421439683+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     watcher event   {"event": "WRITE         \"cert.pem\""}
2025-09-23T10:40:43.421581968+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     modified file   {"name": "cert.pem"}
2025-09-23T10:40:43.422861636+02:00     INFO    webhook/start.go:214    cetificate reloaded
2025-09-23T10:40:43.422970204+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     watcher event   {"event": "WRITE         \"cert.pem\""}
2025-09-23T10:40:43.423190614+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     modified file   {"name": "cert.pem"}
2025-09-23T10:40:53.94114971+02:00      INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     watcher event   {"event": "WRITE         \"key.pem\""}
2025-09-23T10:40:53.941487987+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     modified file   {"name": "key.pem"}
2025-09-23T10:40:53.942776899+02:00     ERROR   sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     failed to reload certificate    {"error": "tls: private key does not match public key"}
panic: failed to reload certificate

goroutine 1 [running]:
main.runStartCmd(0xc0000fc900?, {0x39a89e0?, 0x4?, 0x39a89e4?})
        /home/apanatto/dev/github.com/k8snetworkplumbingwg/sriov-network-operator/cmd/webhook/start.go:216 +0x6be
github.com/spf13/cobra.(*Command).execute(0x5527560, {0xc0007b3dc0, 0x4, 0x4})
        /home/apanatto/go/pkg/mod/github.com/spf13/cobra@v1.8.1/command.go:989 +0xabb
github.com/spf13/cobra.(*Command).ExecuteC(0x5527280)
        /home/apanatto/go/pkg/mod/github.com/spf13/cobra@v1.8.1/command.go:1117 +0x44f
github.com/spf13/cobra.(*Command).Execute(...)
        /home/apanatto/go/pkg/mod/github.com/spf13/cobra@v1.8.1/command.go:1041
main.main()
        /home/apanatto/dev/github.com/k8snetworkplumbingwg/sriov-network-operator/cmd/webhook/main.go:31 +0x1a
exit status 2




------------------------------------------------------------------------

openssl req -x509 -newkey rsa:4096 -keyout key.pem.tmp -out cert.pem.tmp -sha256 -days 3650 -nodes -subj "/C=XX/ST=StateName/L=CityName/O=CompanyName/OU=CompanySectionName/CN=CommonNameOrHostname"; mv cert.pem.tmp cert.pem; mv key.pem.tmp key.pem 

NAMESPACE=openshift-sriov-network-operator go run ./cmd/webhook start --tls-cert-file=cert.pem --tls-private-key-file=key.pem --port=6443 --alsologtostderr=true
2025-09-23T10:42:13.66001923+02:00      INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     Run sriov-network-operator-webhook
2025-09-23T10:42:14.225591018+02:00     ERROR   sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     failed to retrieve supported NICs       {"error": "configmaps \"supported-nic-ids\" not found"}
2025-09-23T10:42:14.226905436+02:00     INFO    sriov-network-operator-webhook  runtime/asm_amd64.s:1700        start server
2025-09-23T10:42:56.167036208+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     watcher event   {"event": "CHMOD         \"cert.pem\""}
2025-09-23T10:42:56.167283176+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     modified file   {"name": "cert.pem"}
2025-09-23T10:42:56.167379563+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     watcher event   {"event": "REMOVE        \"cert.pem\""}
2025-09-23T10:42:56.167403166+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     modified file   {"name": "cert.pem"}
2025-09-23T10:42:56.172092278+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     watcher event   {"event": "CHMOD         \"key.pem\""}
2025-09-23T10:42:56.172173632+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     modified file   {"name": "key.pem"}
2025-09-23T10:42:56.17304502+02:00      INFO    webhook/start.go:214    cetificate reloaded
2025-09-23T10:42:56.173077789+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     watcher event   {"event": "REMOVE        \"key.pem\""}
2025-09-23T10:42:56.17313003+02:00      INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     modified file   {"name": "key.pem"}
2025-09-23T10:43:21.355164878+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     watcher event   {"event": "CHMOD         \"cert.pem\""}
2025-09-23T10:43:21.355237691+02:00     INFO    sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     modified file   {"name": "cert.pem"}
2025-09-23T10:43:21.356006141+02:00     ERROR   sriov-network-operator-webhook  cobra@v1.8.1/command.go:989     failed to reload certificate    {"error": "tls: private key does not match public key"}
panic: failed to reload certificate

goroutine 1 [running]:
main.runStartCmd(0xc000220900?, {0x39a89e0?, 0x4?, 0x39a89e4?})
        /home/apanatto/dev/github.com/k8snetworkplumbingwg/sriov-network-operator/cmd/webhook/start.go:216 +0x6be
github.com/spf13/cobra.(*Command).execute(0x5527560, {0xc000a88d40, 0x4, 0x4})
        /home/apanatto/go/pkg/mod/github.com/spf13/cobra@v1.8.1/command.go:989 +0xabb
github.com/spf13/cobra.(*Command).ExecuteC(0x5527280)
        /home/apanatto/go/pkg/mod/github.com/spf13/cobra@v1.8.1/command.go:1117 +0x44f
github.com/spf13/cobra.(*Command).Execute(...)
        /home/apanatto/go/pkg/mod/github.com/spf13/cobra@v1.8.1/command.go:1041
main.main()
        /home/apanatto/dev/github.com/k8snetworkplumbingwg/sriov-network-operator/cmd/webhook/main.go:31 +0x1a
exit status 2

