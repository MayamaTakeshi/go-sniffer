package core

import (
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"plugin"

	http "go-sniffer/plugSrc/http/build"
	kafka "go-sniffer/plugSrc/kafka/build"
	mongodb "go-sniffer/plugSrc/mongodb/build"
	mssql "go-sniffer/plugSrc/mssql/build"
	mysql "go-sniffer/plugSrc/mysql/build"
	redis "go-sniffer/plugSrc/redis/build"

	"github.com/google/gopacket"
)

type Plug struct {
	dir           string
	ResolveStream func(net gopacket.Flow, transport gopacket.Flow, r io.Reader)
	BPF           string

	InternalPlugList map[string]PlugInterface
	ExternalPlugList map[string]ExternalPlug
}

// All internal plug-ins must implement this interface
// ResolvePacket - entry
// BPFFilter     - set BPF, like: mysql(tcp and port 3306)
// SetFlag       - plug-in params
// Version       - plug-in version
type PlugInterface interface {
	//解析流
	ResolveStream(net gopacket.Flow, transport gopacket.Flow, r io.Reader)
	//BPF
	BPFFilter() string
	//设置插件需要的参数
	SetFlag([]string)
	//获取版本
	Version() string
}

//外部插件
type ExternalPlug struct {
	Name          string
	Version       string
	ResolvePacket func(net gopacket.Flow, transport gopacket.Flow, r io.Reader)
	BPFFilter     func() string
	SetFlag       func([]string)
}

//实例化
func NewPlug() *Plug {

	var p Plug

	//设置默认插件目录
	p.dir, _ = filepath.Abs("./plug/")

	//加载内部插件
	p.LoadInternalPlugList()

	//加载外部插件
	p.LoadExternalPlugList()

	return &p
}

//加载内部插件
func (p *Plug) LoadInternalPlugList() {

	list := make(map[string]PlugInterface)

	//Mysql
	list["mysql"] = mysql.NewInstance()

	//Mongodb
	list["mongodb"] = mongodb.NewInstance()

	//kafka
	list["kafka"] = kafka.NewInstance()

	//Redis
	list["redis"] = redis.NewInstance()

	//Http
	list["http"] = http.NewInstance()

	list["mssql"] = mssql.NewInstance()

	p.InternalPlugList = list
}

//加载外部so后缀插件
func (p *Plug) LoadExternalPlugList() {

	dir, err := ioutil.ReadDir(p.dir)
	if err != nil {
		panic(p.dir + "不存在，或者无权访问")
	}

	p.ExternalPlugList = make(map[string]ExternalPlug)
	for _, fi := range dir {
		if fi.IsDir() || path.Ext(fi.Name()) != ".so" {
			continue
		}

		plug, err := plugin.Open(p.dir + "/" + fi.Name())
		if err != nil {
			panic(err)
		}

		versionFunc, err := plug.Lookup("Version")
		if err != nil {
			panic(err)
		}

		setFlagFunc, err := plug.Lookup("SetFlag")
		if err != nil {
			panic(err)
		}

		BPFFilterFunc, err := plug.Lookup("BPFFilter")
		if err != nil {
			panic(err)
		}

		ResolvePacketFunc, err := plug.Lookup("ResolvePacket")
		if err != nil {
			panic(err)
		}

		version := versionFunc.(func() string)()
		p.ExternalPlugList[fi.Name()] = ExternalPlug{
			ResolvePacket: ResolvePacketFunc.(func(net gopacket.Flow, transport gopacket.Flow, r io.Reader)),
			SetFlag:       setFlagFunc.(func([]string)),
			BPFFilter:     BPFFilterFunc.(func() string),
			Version:       version,
			Name:          fi.Name(),
		}
	}
}

//改变插件地址
func (p *Plug) ChangePath(dir string) {
	p.dir = dir
}

//打印插件列表
func (p *Plug) PrintList() {

	//Print Internal Plug
	for inPlugName, _ := range p.InternalPlugList {
		fmt.Println("internal plug : " + inPlugName)
	}

	//split
	fmt.Println("-- --- --")

	//print External Plug
	for exPlugName, _ := range p.ExternalPlugList {
		fmt.Println("external plug : " + exPlugName)
	}
}

//选择当前使用的插件 && 加载插件
func (p *Plug) SetOption(plugName string, plugParams []string) {

	//Load Internal Plug
	if internalPlug, ok := p.InternalPlugList[plugName]; ok {

		p.ResolveStream = internalPlug.ResolveStream
		internalPlug.SetFlag(plugParams)
		p.BPF = internalPlug.BPFFilter()

		return
	}

	//Load External Plug
	plug, err := plugin.Open("./plug/" + plugName)
	if err != nil {
		panic(err)
	}
	resolvePacket, err := plug.Lookup("ResolvePacket")
	if err != nil {
		panic(err)
	}
	setFlag, err := plug.Lookup("SetFlag")
	if err != nil {
		panic(err)
	}
	BPFFilter, err := plug.Lookup("BPFFilter")
	if err != nil {
		panic(err)
	}
	p.ResolveStream = resolvePacket.(func(net gopacket.Flow, transport gopacket.Flow, r io.Reader))
	setFlag.(func([]string))(plugParams)
	p.BPF = BPFFilter.(func() string)()
}
