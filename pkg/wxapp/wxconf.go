package wxapp

type WxCacheConf struct {
	Addr     string
	Password string
	DbIndex  int
}
type LogConf struct {
	File  string
	Level string
}
type WxConf struct {
	AppId     string
	Secret    string
	Log       LogConf
	HttpDebug bool
	CacheConf WxCacheConf
}
