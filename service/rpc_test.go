package service

//import (
//	"context"
//	"testing"
//	"time"
//
//	wrappers "github.com/golang/protobuf/ptypes/wrappers"
//	"github.com/sirupsen/logrus"
//	v1 "github.com/videocoin/cloud-api/miners/v1"
//	"google.golang.org/grpc"
//)
//
//func TestRPCServer(t *testing.T) {
//	log := logrus.New()
//	logger := log.WithField("version", "test")
//
//	addr := "127.0.0.1:5090"
//
//	opts := &RPCServerOptions{
//		Addr:   addr,
//		DBURI:  "root:@tcp(127.0.0.1:3306)/videocoin?charset=utf8&parseTime=True&loc=Local",
//		Logger: logger,
//	}
//
//	ds, err := NewDatastore(opts.DBURI)
//	checkFail(t, err)
//
//	server, err := NewRPCServer(opts, ds)
//	checkFail(t, err)
//
//	go func() {
//		err := server.Start()
//		checkFail(t, err)
//	}()
//
//	time.Sleep(time.Millisecond * 500)
//
//	cc, err := grpc.Dial(addr, grpc.WithInsecure())
//	checkFail(t, err)
//	defer cc.Close()
//
//	c := v1.NewMinersServiceClient(cc)
//	ctx := context.Background()
//
//	listReq := &v1.ListRequest{}
//	listResp, err := c.List(ctx, listReq)
//	checkFail(t, err)
//
//	// logger.Infof("list: %+v", listResp)
//
//	detailReq := &v1.Request{Id: listResp.Items[0].Id}
//	detailResp, err := c.Get(ctx, detailReq)
//	checkFail(t, err)
//
//	if listResp.Items[0].Id != detailResp.Id {
//		t.Error("failed to get miner detail")
//		t.FailNow()
//	}
//
//	// logger.Infof("detail: %+v", detailResp)
//
//	busyReq := &v1.Request{Id: listResp.Items[0].Id}
//	busyResp, err := c.MarkAsBusy(ctx, busyReq)
//	checkFail(t, err)
//
//	// logger.Infof("busy: %+v", busyResp)
//
//	if busyResp.IsBusy != true {
//		t.Error("failed to mark as busy")
//	}
//
//	idleReq := &v1.Request{Id: listResp.Items[0].Id}
//	idleResp, err := c.MarkAsIdle(ctx, idleReq)
//	checkFail(t, err)
//
//	// logger.Infof("idle: %+v", idleResp)
//
//	if idleResp.IsBusy != false {
//		t.Error("failed to mark as idle")
//	}
//
//	listBusyReq := &v1.ListRequest{IsBusy: &wrappers.BoolValue{Value: true}}
//	listBusyResp, err := c.List(ctx, listBusyReq)
//	checkFail(t, err)
//
//	logger.Infof("list busy: %+v", listBusyResp)
//
//	listIdleReq := &v1.ListRequest{IsBusy: &wrappers.BoolValue{Value: false}}
//	listIdleResp, err := c.List(ctx, listIdleReq)
//	checkFail(t, err)
//
//	logger.Infof("list idle: %+v", listIdleResp)
//}
//
//func checkFail(t *testing.T, err error) {
//	if err != nil {
//		t.Error(err)
//		t.FailNow()
//	}
//}
