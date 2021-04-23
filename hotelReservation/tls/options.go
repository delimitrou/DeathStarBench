package tls

import  (
    "fmt"
    "log"
    "os"
    "strings"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)


var (
    dialopt grpc.DialOption
    serveropt grpc.ServerOption
)


func isTls() bool {
    if val, ok := os.LookupEnv("TLS"); ok {
        if ( strings.EqualFold(val, "false") || strings.EqualFold(val, "0") ) {
            return false
        }else {
            return true
        }
    }
    return false
}


func init() {
    if (isTls()) {
        log.Println("TLS enabled")
        creds, err := credentials.NewClientTLSFromFile("x509/ca_cert.pem", "x.test.example.com")
        if err != nil {
            panic(fmt.Sprintf("failed to load credentials: %v", err))
	}
	dialopt = grpc.WithTransportCredentials(creds)

        creds, err = credentials.NewServerTLSFromFile("x509/server_cert.pem", "x509/server_key.pem")
        if err != nil {
            panic(fmt.Sprintf("failed to create credentials: %v", err))
        }
        serveropt = grpc.Creds(creds)
    } else {
        log.Println("TLS disabled")
        dialopt = nil
        serveropt = nil
    }
}


func GetDialOpt() grpc.DialOption {
    return dialopt
}


func GetServerOpt() grpc.ServerOption {
    return serveropt
}
