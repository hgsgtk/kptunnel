package main

import "flag"
import "fmt"
import "regexp"
import "os"
import "strings"
import "strconv"
import "net/url"

// 2byte の MAX。
// ここを大きくする場合は、WriteItem, ReadItem の処理を変更する。
const BUFSIZE=65535

func hostname2HostInfo( name string ) *HostInfo {
    if strings.Index( name, "://" ) == -1 {
        name = fmt.Sprintf( "http://%s", name )
    }
    serverUrl, err := url.Parse( name )
    if err != nil {
        fmt.Printf( "%s\n", err )
        return nil
    }
    hostport := strings.Split( serverUrl.Host, ":" )
    if len( hostport ) != 2 {
        fmt.Printf( "illegal pattern. set 'hoge.com:1234'\n" )
        return nil
    }
    var port int
    port, err2 := strconv.Atoi( hostport[1] )
    if err2 != nil {
        fmt.Printf( "%s\n", err2 )
        return nil
    }
    return &HostInfo{ "", hostport[ 0 ], port, serverUrl.Path }
}

func main() {

    var cmd = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
    mode := cmd.String( "mode", "",
        "<server|r-server|wsserver|r-wsserver|client|r-client|wsclient|r-wsclient>" )
    server := cmd.String( "server", "", "server (hoge.com:1234 or :1234)" )
    remote := cmd.String( "remote", "", "remote host (hoge.com:1234)" )
    pass := cmd.String( "pass", "hogehoge", "password" )
    encPass := cmd.String( "encPass", "hogehoge", "packet encrypt pass" )
    encCount := cmd.Int( "encCount", 1000,
        `number to encrypt the tunnel packet.
 -1: infinity
  0: plain
  N: packet count` )
    ipPattern := cmd.String( "ip", "", "allow ip pattern" )
    proxyHost := cmd.String( "proxy", "", "proxy server" )
    userAgent := cmd.String( "UA", "Go Http Client", "user agent for websocket" )
    sessionPort := cmd.Int( "port", 0, "session port" )

    usage := func() {
        fmt.Fprintf(cmd.Output(), "\nUsage: %s [options]\n\n", os.Args[0])
        fmt.Fprintf(cmd.Output(), " options:\n" )
        cmd.PrintDefaults()
        os.Exit( 1 )
    }
    cmd.Usage = usage

    cmd.Parse( os.Args[1:] )


    var remoteInfo *HostInfo
    if *remote != "" {
        remoteInfo = hostname2HostInfo( *remote )
    }
    
    serverInfo := hostname2HostInfo( *server )
    if serverInfo == nil {
        fmt.Print( "set -server option!\n" )
        usage()
    }

    if *mode == "r-server" || *mode == "r-wsserver" ||
        *mode == "client" || *mode == "wsclient" {
        if *sessionPort == 0 {
            fmt.Print( "set -port option!\n" )
            usage()
        }
        if remoteInfo == nil {
            fmt.Print( "set -remote option!\n" )
            usage()
        }
    }
    
    
    echoPort := 8002
    websocketServerInfo := HostInfo{ "ws://", serverInfo.Name, serverInfo.Port, "/" }
    var pattern *regexp.Regexp
    if *ipPattern != "" {
        pattern = regexp.MustCompile( *ipPattern )
    }
    if *pass == "" {
        pass = nil
    }

    param := &TunnelParam{ pass, *mode, pattern, encPass, *encCount }

    switch *mode {
    case "server":
        StartServer( param, serverInfo.Port )
    case "r-server":
        StartReverseServer( param, serverInfo.Port, *sessionPort, *remoteInfo )
    case "wsserver":
        StartWebsocketServer( param, serverInfo.Port )
    case "r-wsserver":
        StartReverseWebSocketServer( param, serverInfo.Port, *sessionPort, *remoteInfo )
    case "client":
        StartClient( param, *serverInfo, *sessionPort, *remoteInfo )
    case "r-client":
        StartReverseClient( param, *serverInfo )
    case "wsclient":
        StartWebSocketClient( *userAgent, param, websocketServerInfo, *proxyHost, *sessionPort, *remoteInfo )
    case "r-wsclient":
        StartReverseWebSocketClient( *userAgent, param, websocketServerInfo, *proxyHost )
    case "echo":
        StartEchoServer( echoPort )
    case "test":
        for index := 0; index < 2; index++ {
            val := make([]byte,10)
            fmt.Printf( "%p\n", val )
        }
            
    }
}
