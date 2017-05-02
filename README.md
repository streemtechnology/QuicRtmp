# QuicRtmp
Rtmp with Quic based transport medium


## Dependency installation

go get github.com/zhangpeihao/gortmp
go get github.com/marten-seemann/quic-conn
go get github.com/lucas-clemente/quic-go


## To Build

go build rtmp_srv.go

go build rtmp_client.go

## To run 

export GOPATH=one dir above pwd

./rtmp_srv -s

./rtmp_client -c

