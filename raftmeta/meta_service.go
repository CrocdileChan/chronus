package raftmeta

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/angopher/chronus/raftmeta/internal"
	"github.com/angopher/chronus/services/meta"
	"github.com/angopher/chronus/x"
	"github.com/influxdata/influxql"
	"io/ioutil"
	"net/http"
	"time"
)

type CommonResp struct {
	RetCode int    `json:ret_code`
	RetMsg  string `json:ret_msg`
}

type MetaService struct {
	node          *RaftNode
	cli           *meta.Client
	Linearizabler interface {
		ReadNotify(ctx context.Context) error
	}
}

func NewMetaService(cli *meta.Client, node *RaftNode, l *Linearizabler) *MetaService {
	return &MetaService{
		cli:           cli,
		node:          node,
		Linearizabler: l,
	}
}

func (s *MetaService) ProposeAndWait(msgType int, data []byte, retData interface{}) error {
	timeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	pr := &internal.Proposal{Type: msgType}
	pr.Data = data

	resCh := make(chan error)
	go func() {
		err := s.node.ProposeAndWait(ctx, pr, retData)
		resCh <- err
	}()

	var err error
	select {
	case err = <-resCh:
	}

	return err
}

type CreateDatabaseReq struct {
	Name string
}

type CreateDatabaseResp struct {
	CommonResp
	DbInfo meta.DatabaseInfo
}

func (s *MetaService) CreateDatabase(w http.ResponseWriter, r *http.Request) {
	resp := new(CreateDatabaseResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req CreateDatabaseReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	db := &meta.DatabaseInfo{}
	err = s.ProposeAndWait(internal.CreateDatabase, data, db)
	if err == nil {
		resp.DbInfo = *db
		resp.RetCode = 0
		resp.RetMsg = "ok"
	} else {
		resp.RetMsg = fmt.Sprintf("msg=create database failed,database=%s,err_msg=%v", req.Name, err)
	}
	return
}

type DropDatabaseReq struct {
	Name string
}

type DropDatabaseResp struct {
	CommonResp
}

func (s *MetaService) DropDatabase(w http.ResponseWriter, r *http.Request) {
	resp := new(DropDatabaseResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req DropDatabaseReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.DropDatabase, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
	return
}

type DropRetentionPolicyReq struct {
	Database string
	Policy   string
}

type DropRetentionPolicyResp struct {
	CommonResp
}

func (s *MetaService) DropRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	resp := new(DropRetentionPolicyResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req DropRetentionPolicyReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.DropRetentionPolicy, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		fmt.Printf("msg=DropRetetionPolicy failed,database=%s,policy=%s,err_msg=%s\n", req.Database, req.Policy, err.Error())
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
	return
}

type CreateShardGroupReq struct {
	Database  string
	Policy    string
	Timestamp int64
}

type CreateShardGroupResp struct {
	CommonResp
	ShardGroupInfo meta.ShardGroupInfo
}

func (s *MetaService) CreateShardGroup(w http.ResponseWriter, r *http.Request) {
	resp := new(CreateShardGroupResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req CreateShardGroupReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	sg := &meta.ShardGroupInfo{}
	err = s.ProposeAndWait(internal.CreateShardGroup, data, sg)
	if err != nil {
		resp.RetMsg = err.Error()
		fmt.Printf("msg=RetetionPolicy failed,database=%s,policy=%s,err_msg=%s", req.Database, req.Policy, err.Error())
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
	resp.ShardGroupInfo = *sg
	return
}

type CreateDataNodeReq struct {
	HttpAddr string
	TcpAddr  string
}
type CreateDataNodeResp struct {
	CommonResp
	NodeInfo meta.NodeInfo
}

func (s *MetaService) CreateDataNode(w http.ResponseWriter, r *http.Request) {
	fmt.Println("create data node")
	resp := new(CreateDataNodeResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req CreateDataNodeReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	ni := &meta.NodeInfo{}
	err = s.ProposeAndWait(internal.CreateDataNode, data, ni)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
	resp.NodeInfo = *ni
	fmt.Println("create data node done")
}

type DeleteDataNodeReq struct {
	Id uint64
}
type DeleteDataNodeResp struct {
	CommonResp
}

func (s *MetaService) DeleteDataNode(w http.ResponseWriter, r *http.Request) {
	fmt.Println("create data node")
	resp := new(DeleteDataNodeResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req DeleteDataNodeReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.DeleteDataNode, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type RetentionPolicySpec struct {
	Name               string
	ReplicaN           int
	Duration           time.Duration
	ShardGroupDuration time.Duration
}

type CreateRetentionPolicyReq struct {
	Database    string
	Rps         RetentionPolicySpec
	MakeDefault bool
}
type CreateRetentionPolicyResp struct {
	CommonResp
	RetentionPolicyInfo meta.RetentionPolicyInfo
}

func (s *MetaService) CreateRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreateRetentionPolicy")
	resp := new(CreateRetentionPolicyResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req CreateRetentionPolicyReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	rpi := &meta.RetentionPolicyInfo{}
	err = s.ProposeAndWait(internal.CreateRetentionPolicy, data, rpi)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
	resp.RetentionPolicyInfo = *rpi
}

type UpdateRetentionPolicyReq struct {
	Database    string
	Name        string
	Rps         RetentionPolicySpec
	MakeDefault bool
}
type UpdateRetentionPolicyResp struct {
	CommonResp
}

func (s *MetaService) UpdateRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	resp := new(UpdateRetentionPolicyResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req UpdateRetentionPolicyReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.UpdateRetentionPolicy, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type CreateDatabaseWithRetentionPolicyReq struct {
	Name string
	Rps  RetentionPolicySpec
}
type CreateDatabaseWithRetentionPolicyResp struct {
	CommonResp
	DbInfo meta.DatabaseInfo
}

func (s *MetaService) CreateDatabaseWithRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	resp := new(CreateDatabaseWithRetentionPolicyResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req CreateDatabaseWithRetentionPolicyReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	db := &meta.DatabaseInfo{}
	err = s.ProposeAndWait(internal.CreateDatabaseWithRetentionPolicy, data, db)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.DbInfo = *db
	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type CreateUserReq struct {
	Name     string
	Password string
	Admin    bool
}
type CreateUserResp struct {
	CommonResp
	UserInfo meta.UserInfo
}

func (s *MetaService) CreateUser(w http.ResponseWriter, r *http.Request) {
	resp := new(CreateUserResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req CreateUserReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	user := &meta.UserInfo{}
	err = s.ProposeAndWait(internal.CreateUser, data, user)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.UserInfo = *user
	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type DropUserReq struct {
	Name string
}
type DropUserResp struct {
	CommonResp
}

func (s *MetaService) DropUser(w http.ResponseWriter, r *http.Request) {
	resp := new(DropUserResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req DropUserReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.DropUser, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type UpdateUserReq struct {
	Name     string
	Password string
}
type UpdateUserResp struct {
	CommonResp
}

func (s *MetaService) UpdateUser(w http.ResponseWriter, r *http.Request) {
	resp := new(UpdateUserResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req UpdateUserReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.UpdateUser, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type SetPrivilegeReq struct {
	UserName  string
	Database  string
	Privilege influxql.Privilege
}
type SetPrivilegeResp struct {
	CommonResp
}

func (s *MetaService) SetPrivilege(w http.ResponseWriter, r *http.Request) {
	resp := new(SetPrivilegeResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req SetPrivilegeReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.SetPrivilege, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type SetAdminPrivilegeReq struct {
	UserName string
	Admin    bool
}
type SetAdminPrivilegeResp struct {
	CommonResp
}

func (s *MetaService) SetAdminPrivilege(w http.ResponseWriter, r *http.Request) {
	resp := new(SetAdminPrivilegeResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req SetAdminPrivilegeReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.SetAdminPrivilege, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type AuthenticateReq struct {
	UserName string
	Password string
}
type AuthenticateResp struct {
	CommonResp
	UserInfo meta.UserInfo
}

func (s *MetaService) Authenticate(w http.ResponseWriter, r *http.Request) {
	resp := new(AuthenticateResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req AuthenticateReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	user := &meta.UserInfo{}
	err = s.ProposeAndWait(internal.Authenticate, data, user)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.UserInfo = *(user)
	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type DropShardReq struct {
	Id uint64
}
type DropShardResp struct {
	CommonResp
}

func (s *MetaService) DropShard(w http.ResponseWriter, r *http.Request) {
	resp := new(DropShardResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req DropShardReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.DropShard, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type TruncateShardGroupsReq struct {
	Time time.Time
}
type TruncateShardGroupsResp struct {
	CommonResp
}

func (s *MetaService) TruncateShardGroups(w http.ResponseWriter, r *http.Request) {
	resp := new(TruncateShardGroupsResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req TruncateShardGroupsReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.TruncateShardGroups, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type PruneShardGroupsResp struct {
	CommonResp
}

func (s *MetaService) PruneShardGroups(w http.ResponseWriter, r *http.Request) {
	resp := new(PruneShardGroupsResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	err := s.ProposeAndWait(internal.PruneShardGroups, []byte{}, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

//DeleteShardGroup
type DeleteShardGroupReq struct {
	Database string
	Policy   string
	Id       uint64
}
type DeleteShardGroupResp struct {
	CommonResp
}

func (s *MetaService) DeleteShardGroup(w http.ResponseWriter, r *http.Request) {
	resp := new(DeleteShardGroupResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req DeleteShardGroupReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.DeleteShardGroup, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

//PrecreateShardGroups
type PrecreateShardGroupsReq struct {
	From time.Time
	To   time.Time
}
type PrecreateShardGroupsResp struct {
	CommonResp
}

func (s *MetaService) PrecreateShardGroups(w http.ResponseWriter, r *http.Request) {
	resp := new(PrecreateShardGroupsResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req PrecreateShardGroupsReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.PrecreateShardGroups, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

//CreateContinuousQuery
type CreateContinuousQueryReq struct {
	Database string
	Name     string
	Query    string
}
type CreateContinuousQueryResp struct {
	CommonResp
}

func (s *MetaService) CreateContinuousQuery(w http.ResponseWriter, r *http.Request) {
	resp := new(CreateContinuousQueryResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req CreateContinuousQueryReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.CreateContinuousQuery, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

//DropContinuousQuery
type DropContinuousQueryReq struct {
	Database string
	Name     string
}
type DropContinuousQueryResp struct {
	CommonResp
}

func (s *MetaService) DropContinuousQuery(w http.ResponseWriter, r *http.Request) {
	resp := new(DropContinuousQueryResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req DropContinuousQueryReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.DropContinuousQuery, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

//CreateSubscription
type CreateSubscriptionReq struct {
	Database     string
	Rp           string
	Name         string
	Mode         string
	Destinations []string
}
type CreateSubscriptionResp struct {
	CommonResp
}

func (s *MetaService) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	resp := new(CreateSubscriptionResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req CreateSubscriptionReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.CreateSubscription, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

//DropSubscription
type DropSubscriptionReq struct {
	Database string
	Rp       string
	Name     string
}
type DropSubscriptionResp struct {
	CommonResp
}

func (s *MetaService) DropSubscription(w http.ResponseWriter, r *http.Request) {
	resp := new(DropSubscriptionResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req DropSubscriptionReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	err = s.ProposeAndWait(internal.DropSubscription, data, nil)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type AcquireLeaseReq struct {
	Name   string
	NodeId uint64
}

type AcquireLeaseResp struct {
	CommonResp
	Lease meta.Lease
}

func (s *MetaService) AcquireLease(w http.ResponseWriter, r *http.Request) {
	resp := new(AcquireLeaseResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	data, err := ioutil.ReadAll(r.Body)
	x.Check(err)

	var req AcquireLeaseReq
	if err := json.Unmarshal(data, &req); err != nil {
		resp.RetMsg = err.Error()
		return
	}

	lease := &meta.Lease{}
	err = s.ProposeAndWait(internal.AcquireLease, data, lease)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.Lease = *lease
	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type DataResp struct {
	CommonResp
	Data []byte
}

func (s *MetaService) Data(w http.ResponseWriter, r *http.Request) {
	resp := new(DataResp)
	resp.RetCode = -1
	resp.RetMsg = "fail"
	defer WriteResp(w, &resp)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := s.Linearizabler.ReadNotify(ctx)
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	data := s.cli.Data()
	resp.Data, err = data.MarshalBinary()
	if err != nil {
		resp.RetMsg = err.Error()
		return
	}

	resp.RetCode = 0
	resp.RetMsg = "ok"
}

type PingResp struct {
	CommonResp
	Index uint64
}

func (s *MetaService) Ping(w http.ResponseWriter, r *http.Request) {
	resp := new(PingResp)
	resp.Index = s.cli.Data().Index
	resp.RetCode = 0
	resp.RetMsg = "ok"
	WriteResp(w, &resp)
}

func WriteResp(w http.ResponseWriter, v interface{}) error {
	bytes, _ := json.Marshal(v)
	_, err := w.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
