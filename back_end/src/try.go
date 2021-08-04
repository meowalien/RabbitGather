package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/smtp"
	"rabbit_gather/src/db_operator"
	"rabbit_gather/util"
)

type ABC string

func (a ABC) name() string {
	return string(a)
}

// API list
type PermissionCode uint32
type APIPermissionCode uint32

const (
	OK                              = APIPermissionCode(^uint32(0))
	Admin                           = APIPermissionCode(uint32(0))
	SearchArticle APIPermissionCode = 1 << iota // token is malformed
	PostArticle
	Login
	GetPeerID
	PeerWebsockt
	PeerWebsocketHandler
	GetPeerIDHandler
	UserAccountLoginHandler
	PermissionCheck
)

//const (
//	Ldate         = 1 << iota     // the date in the local time zone: 2009/01/23
//	Ltime                         // the time in the local time zone: 01:23:23
//	Lmicroseconds                 // microsecond resolution: 01:23:23.123123.  assumes Ltime.
//	Llongfile                     // full file name and line number: /a/b/c/d.go:23
//	Lshortfile                    // final file name element and line number: d.go:23. overrides Llongfile
//	LUTC                          // if Ldate or Ltime is set, use UTC rather than the local time zone
//	Lmsgprefix                    // move the "prefix" from the beginning of the line to before the message
//	LstdFlags     = Ldate | Ltime // initial values for the standard logger
//)

const (
	DEBUG uint8 = 1 << iota
	WARNING
	ERROR
)
const (
	MUTE = uint8(0)
	ALL  = ^uint8(0)
)

var LogLevelMask = ALL

// Foreground text colors
// Base attributes
//const EndColor = "\033[00m"
//
//func ColorSting(s string ,color ColorCode) string{
//	return fmt.Sprintf("\033[%dm%s\033[00m",color,s,EndColor)
//}
type Msg struct {
	UserId    string   `json:"userId"`
	Text      string   `json:"text"`
	State     string   `json:"state"`
	Namespace string   `json:"namespace"`
	Rooms     []string `json:"rooms"`
}
type NameStruct struct {
	Name string
	Num  int
}
type UserClaims struct {
	User string
}
type UserC UserClaims
type Claims interface {
}

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unkown fromServer")
		}
	}
	return nil, nil
}

var dbOperator db_operator.DBOperator

func init() {
	//log = logger.NewLoggerWrapper("AccountManagement")
	type Config struct {
		DatabaseConfig db_operator.DatabaseConnectionConfiguration `json:"database_config"`
	}
	var config Config
	err := util.ParseJsonConfic(&config, "config/article_management.config.json")
	if err != nil {
		panic(err.Error())
	}
	dbOperator = db_operator.GetOperator(db_operator.Mysql, config.DatabaseConfig)
	//fmt.Println("")
}

type ArticleAuthoritySetting struct {
	MaxRadius uint32 `json:"max_radius"`
	MinRadius uint32 `json:"min_radius"`
}

func (m ArticleAuthoritySetting) Value() (driver.Value, error) {
	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return driver.Value(j), nil
}

func (m *ArticleAuthoritySetting) Scan(src interface{}) error {
	fmt.Println("src: ", src)
	var source []byte

	switch src.(type) {
	case []uint8:
		source = src.([]byte)
	case nil:
		return nil
	default:
		return errors.New("incompatible type for StringInterfaceMap")
	}
	fmt.Println("source: ", string(source))
	var _m ArticleAuthoritySetting
	err := json.Unmarshal(source, &_m)
	if err != nil {
		return err
	}
	fmt.Println("_m: ", _m)
	*m = _m
	return nil
}

//var log = logger.NewLoggerWrapper("TRY").TempLog()
func main() {
	stat := dbOperator.Statement("select setting from `user_article_setting` where user = ?;")
	var setting ArticleAuthoritySetting
	err := stat.QueryRow(1).Scan(&setting)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(setting)
	//err := redis_db.GetClient(0).IteratorAllKeyValue(context.Background(),
	//	func(_ redis.Pipeliner, key string, value interface{}) (bool, error) {
	//		fmt.Println(key," ",value)
	//		return true,nil
	//	})
	//if err != nil {
	//	panic(err.Error())
	//}

	//claims := Claims(&UserClaims{User: "USER"})
	//var claimstruct UserClaims
	//var interfaceclaims interface{} = &claimstruct
	//*claims.(*UserClaims) =*interfaceclaims.(*UserClaims)
	//interfaceclaims.(*UserClaims) = claims.(*UserClaims)
	//fmt.Println(*claims.(*UserClaims))
	//fmt.Println(*interfaceclaims.(*UserClaims))
	//*interfaceclaims.(*UserClaims) = *claims.(*UserClaims)

	//marsual := func(it interface{}){
	//	switch t :=it.(type) {
	//	case *UserClaims:
	//		*t = *claims.(*UserClaims)
	//	}
	//}
	//
	//marsual(&claimstruct)
	//fmt.Println(claimstruct)

	//ctx := context.Background()
	//client := redis_db.GetClient(0)
	//
	//err := client.Set(ctx, "KEY", NameStruct{Name: "THE_NAME", Num: 123}, 0)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//var rss NameStruct
	//err = client.Get(ctx, "KEY", &rss)
	//
	//if err != nil {
	//	fmt.Println(err.Error())
	//} else {
	//	pretty.Println(rss)
	//}
	//client.DeleteAll(ctx, "*")
	//
	////client.Del(ctx,"KEY")
	//err = client.Get(ctx, "KEY", &rss)
	//
	//if err != nil {
	//	fmt.Println(err.Error())
	//} else {
	//	pretty.Println(rss)
	//}
	//pong, err :=  client.Ping(ctx).Result()
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(pong)
	//val, err := client.Get(ctx, "key").Result()
	//switch {
	//case err == redis.Nil:
	//	fmt.Println("key does not exist")
	//case err != nil:
	//	fmt.Println("Get failed", err)
	//case val == "":
	//	fmt.Println("value is empty")
	//}	// Output: PONG <nil>
	//
	//var incr *redis.IntCmd
	//_, err = client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
	//	incr = pipe.Incr(ctx, "pipelined_counter")
	//	pipe.Expire(ctx, "pipelined_counter", time.Hour)
	//	return nil
	//})
	//if err != nil {
	//	panic(err)
	//}
	//
	//// The value is available only after pipeline is executed.
	//fmt.Println(incr.Val())
	//pipe := client.Pipeline()
	//
	//incr = pipe.Incr(ctx, "keypipeline_counter")
	//pipe.Expire(ctx, "keypipeline_counter", time.Hour)
	//
	//_, err = pipe.Exec(ctx)
	//defer pipe.Close()
	//if err != nil {
	//	panic(err)
	//}
	//// The value is available only after Exec.
	//fmt.Println(incr.Val())
	//
	//for i := 0; i < 106; i++ {
	//	client.Set(ctx,"key"+strconv.Itoa(i), "value"+strconv.Itoa(i), 0)
	//}
	//
	//var cursor uint64
	//var n int
	//for {
	//	var keys []string
	//	var err error
	//	//*扫描所有key，每次20条
	//	keys, cursor, err = client.Scan(ctx,cursor, "*", 20).Result()
	//	if err != nil {
	//		panic(err)
	//	}
	//	n += len(keys)
	//
	//	fmt.Printf("\nfound %d keys\n", n)
	//	var value string
	//	for _, key := range keys {
	//		value, err = client.Get(ctx,key).Result()
	//		fmt.Printf("%v %v\n", key, value)
	//	}
	//	if cursor == 0 {
	//		break
	//	}
	//}
	//for {
	//	var keys []string
	//	var err error
	//	//*扫描所有key，每次20条
	//	keys, cursor, err = 	pipe.Scan(ctx,0,"*",20).Result()
	//	if err != nil {
	//		panic(err)
	//	}
	//	n += len(keys)
	//
	//	fmt.Printf("\nfound %d keys\n", n)
	//	var value string
	//	for _, key := range keys {
	//		value, err = client.Get(ctx,key).Result()
	//		fmt.Printf("%v %v\n", key, value)
	//	}
	//	if cursor == 0 {
	//		break
	//	}
	//}
	//iter := client.Scan(ctx, cursor, "key*", 0).IteratorKeyValue()
	//for iter.Next(ctx) {
	//	fmt.Println(iter.Val())
	//}
	//if err := iter.Err(); err != nil {
	//	panic(err)
	//}
	//keys, cursor, err = client.Scan(cursor, "*", 20).Result()

	//rdb.Set(context.Background(),"key","")
	//return nil
	//server := socketio.NewServer(nil)
	////if err != nil {
	////	fmt.Println(err)
	////}
	//server.OnConnect("/", func(s socketio.Conn) error {
	//	msg := Msg{s.ID(), "connected!", "notice", "", nil}
	//	s.SetContext("")
	//	s.Emit("res", msg)
	//	fmt.Println("connected /:", s.ID())
	//	// fmt.Printf("URL: %#v \n", s.URL())
	//	// fmt.Printf("LocalAddr: %#+v \n", s.LocalAddr())
	//	// fmt.Printf("RemoteAddr: %#+v \n", s.RemoteAddr())
	//	// fmt.Printf("RemoteHeader: %#+v \n", s.RemoteHeader())
	//	// fmt.Printf("Cookies: %s \n", s.RemoteHeader().Get("Cookie"))
	//	return nil
	//})
	//
	//server.OnEvent("/", "join", func(s socketio.Conn, room string) {
	//	s.Join(room)
	//	msg := Msg{s.ID(), "<= " + s.ID() + " join " + room, "state", s.Namespace(), s.Rooms()}
	//	fmt.Println("/:join", room, s.Namespace(), s.Rooms())
	//	server.BroadcastToRoom(room, "res", "join", msg)
	//})
	//server.OnEvent("/", "leave", func(s socketio.Conn, room string) {
	//	s.Leave(room)
	//	msg := Msg{s.ID(), "<= " + s.ID() + " leave " + room, "state", s.Namespace(), s.Rooms()}
	//	fmt.Println("/:chat received", room, s.Namespace(), s.Rooms())
	//	server.BroadcastToRoom(room, "res", "leave", msg)
	//})
	//
	//server.OnEvent("/", "chat", func(s socketio.Conn, msg string) {
	//	res := Msg{s.ID(), "<= " + msg, "normal", s.Namespace(), s.Rooms()}
	//	s.SetContext(res)
	//	fmt.Println("/:chat received", msg, s.Namespace(), s.Rooms(), server.Rooms("/"))
	//	rooms := s.Rooms()
	//	if len(rooms) > 0 {
	//		fmt.Println("broadcast to", rooms)
	//		for i := range rooms {
	//			server.BroadcastToRoom(rooms[i], "res", "chat", res)
	//		}
	//		// server.BroadcastToRoom(s.Rooms()[0], "res", res)
	//	}
	//})
	//
	//server.OnEvent("/", "notice", func(s socketio.Conn, msg string) {
	//	fmt.Println("/:notice:", msg)
	//	s.Emit("reply", "have "+msg)
	//})
	//server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
	//	fmt.Println("/chat:msg received", msg)
	//	return "recv " + msg
	//})
	//server.OnEvent("/", "bye", func(s socketio.Conn) string {
	//	last := s.Context().(Msg)
	//	s.Emit("bye", last)
	//	res := Msg{s.ID(), "<= " + s.ID() + " leaved", "state", s.Namespace(), s.Rooms()}
	//	rooms := s.Rooms()
	//	s.LeaveAll()
	//	s.Close()
	//	if len(rooms) > 0 {
	//		fmt.Println("broadcast to", rooms)
	//		for i := range rooms {
	//			server.BroadcastToRoom(rooms[i], "res", "bye", res)
	//		}
	//	}
	//	fmt.Printf("/:bye last context: %#+v \n", s.Context())
	//	return last.Text
	//})
	//
	//server.OnError("/", func(s socketio.Conn, e error) {
	//	fmt.Println("/:error ", e)
	//})
	//server.OnDisconnect("/", func(s socketio.Conn, reason string) {
	//	fmt.Println("/:closed", s.ID(), reason)
	//})
	//
	//go server.Serve()
	//defer server.Close()
	//
	//http.Handle("/socket.io/", server)
	//http.Handle("/", http.FileServer(http.Dir("./asset")))
	//
	//fmt.Println("Serving at localhost:8000...")
	//fmt.Println(http.ListenAndServe(":8000", nil))
	//type PositionStruct struct {
	//	X  float32 `json:"latitude"`
	//	Y float32 `json:"longitude"`
	//}
	//
	//type SearchArticleRequest struct {
	//	Position PositionStruct `json:"position"`
	//	Radius   int            `json:"radius"`
	//}
	//searchArticleRequest := SearchArticleRequest{
	//	Position: PositionStruct{
	//		X:  54.44865417480469,
	//		Y: 145.48887634277344,
	//	},
	//	Radius: 99999999999,
	//}
	//type Props struct {
	//	ID string `json:"id"`
	//	Title string `json:"title"`
	//	Content string `json:"content"`
	//	Timestamp int64 `json:"timestamp"`
	//}
	//type Labels struct {
	//
	//}
	//type Article struct {
	//	Id int
	//	Labels Labels
	//	Props Props
	//}
	//driver := neo4j_db.GetDriver()
	//session := driver.NewSession(neo4j.SessionConfig{})
	//defer session.Close()
	//result, err := session.Run(util.GetFileStoredPlainText("sql/search_article_with_radius.cyp"),
	//	map[string]interface{}{
	//		"longitude": searchArticleRequest.Position.Y,
	//		"latitude":  searchArticleRequest.Position.X,
	//		"radius":    searchArticleRequest.Radius,
	//	},
	//)
	//if err != nil {
	//	panic(err.Error())
	//}
	//for result.Next() {
	//	record := result.Record()
	//	s, _ := record.Get("article")
	//	pretty.Println(s.(neo4j.Node).Props)
	//}
	//res, err := neo4j_db.RunScriptWithScriptFile(
	//	"sql/search_article_with_radius.cyp",
	//	map[string]interface{}{
	//		"longitude":  searchArticleRequest.Position.Y,
	//		"latitude":     searchArticleRequest.Position.X,
	//		"radius":searchArticleRequest.Radius,
	//	})
	//if err != nil {
	//	panic(err.Error())
	//}
	//fmt.Println("res: ",res)
	//
	//for res.Next() {
	//	record := res.Record()
	//fmt.Println("record: ",record)
	//	if value, ok := record.Get("Article"); ok {
	//		fmt.Println(value)
	//	}
	//	if value, ok := record.Get("properties"); ok {
	//		fmt.Println(value)
	//	}
	//}
	//a := claims.StatusClaims{StatusBitmask: status_bitmask.WaitVerificationCode}
	//fmt.Println(a.Valid())

	//m := map[string]int{}
	//fmt.Println(util.NewVerificationCodeWithLength(4))
	//utClaims := auth.UtilityClaims{
	//	auth.StatusClaimsName: auth.StatusClaims{
	//		StatusBitmask: auth.Login,
	//	},
	//}
	//var st auth.StatusClaims
	//utClaims.MappingClaim(&st)
	//pretty.Println(st)
	//fmt.Println(st.StatusBitmask)
	//fmt.Println(utClaims.Valid())

}

//fmt.Println(1|4)
//err :=fmt.Errorf("SentMail: %w", errors.New("ssss"))
//color := "\033[91m"//red

//color := "\033["+strconv.Itoa(FgRed)+"m"
//colodr := "\033[%dm"
//value := "ABC"
//endcolor := "\033[00m"
//c := fmt.Sprintf("\033[%dm%s\033[00m",FgYellow,"THS")
//fmt.Println(util.ColorSting("THISSTRING",util.FgCyan))
//d := 5
//p := fmt.Sprintf("%%%dd", d)
//fmt.Println(p)
//fmt.Printf(fmt.Sprintf("%%%dd", d), util.GetSnowflakeIntWithLength(int64(d)))

//fmt.Println(math.Log10(float64(9)))
//fmt.Println(math.Log10(float64(9999)))
//fmt.Printf("%04d\n", util.GetRandomInt(9, 9999))
//fmt.Sprintf("%d\n",  util.GetSnowflakeIntWithLength(4))
//s := time.Now()
//for i := 0; i < 9999999; i++ {
//	fmt.Sprintf("%d\n",  util.GetSnowflakeIntWithLength(4))
//	//util.GetRandomInt(9, 9999)
//}
//fmt.Println(time.Since(s))
//
//s = time.Now()
//for i := 0; i < 9999999; i++ {
//	fmt.Sprintf("%4d\n",  util.GetSnowflakeIntWithLength(4))
//}
//fmt.Println(time.Since(s))
//target := int64(util.MaxInt*-1)
//fmt.Println(target)
//fmt.Println(target%int(math.Pow(10, float64(3))))
//fmt.Println(util.CutIntMax(target,3))
//min := 2
//max := 5
//if min < 1 {
//	panic("min must >= 1")
//}
//res := target / int(math.Pow(10, float64(min-1))) % int(math.Pow(10, float64(max-min+1)))
//
//fmt.Println(res)

//for i := 1; i <= max-min+1; i++ {
//	a := int(math.Pow(10, float64(i))) // 10
//	app :=   ((tempA % a )/(i*a)) *a
//	fmt.Println("---",(tempA % a ))
//	fmt.Println(app)
//	res += app
//}
//fmt.Println(res)
// 5678

//maxint := ^uint32(0)
//fmt.Printf("%64b\n", DEBUG)
//fmt.Printf("%64b\n", WARNING)
//fmt.Printf("%64b\n", ^uint(0) >>1)
//fmt.Printf("%d\n", int^uint64(0)  >> 1)
//fmt.Printf("%04d",util.GetRandomInt(0,100000))
//fmt.Println(util.Snowflake().Base58())
//fmt.Println(ALL&DEBUG)
//fmt.Println(ALL&WARNING)
//fmt.Println(ALL&ERROR)
//s:="TTTTSSSSS"
//c := color.New(color.FgBlue)
////fmt.Println("5555555555",c(s))
////ll := rawlog.New(os.Stdout,"",rawlog.Lshortfile)
//c.Fprint(os.Stdout,s)
//ll.Printf("%s",c(s))

//log := logger.NewLogger("")
//log.ERROR.Println("Error message.")
//log.WARNING.Println("WARNING message.")
//log.DEBUG.Println("DEBUG message.")
//errorLogger := log.New(os.Stdout,"ERROR: ", log.Ltime|log.Ldate|log.Llongfile|log.Lmsgprefix)
//errorLogger.Println("123ABC")
//warningLogger := log.New(os.Stdout,"WARNING: ", log.Ltime|log.Ldate|log.Lshortfile|log.Lmsgprefix)
//warningLogger.Println("123ABC")
//debugLogger := log.New(os.Stdout,"DEBUG: ", log.Ltime|log.Ldate|log.Lshortfile|log.Lmsgprefix)
//debugLogger.Println("123ABC")
//
//
//loggerInst1 := log.New(os.Stdout,"---", log.Ltime|log.Lshortfile)
//loggerInst1.Println("123ABC")
//fmt.Println(OK)
//fmt.Println(SearchArticle)
////fmt.Println(OK)
////fmt.Println(OK)
////fmt.Println(OK)
//userPermission := SearchArticle | PostArticle
//fmt.Println(userPermission)
//fmt.Println(SearchArticle)
//fmt.Println(PostArticle)
//fmt.Println(Login)
//fmt.Println("----------")
//fmt.Println(SearchArticle & userPermission)
//fmt.Println(PostArticle & userPermission)
//fmt.Println(OK & userPermission)
//fmt.Println(Login & userPermission)     //0
//fmt.Println(GetPeerID & userPermission) //0
//fmt.Println(Admin & userPermission)     //0

//if 1&1{
//}
//var Errors =  uint32(1)
//fmt.Println(1 << 0)
//fmt.Println(jwt.ValidationErrorMalformed )
//fmt.Println(Errors)
//fmt.Println(Errors&jwt.ValidationErrorMalformed )
//fmt.Println(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet )
//fmt.Println(1 &(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) )
//
////fmt.Println(jwt.ValidationErrorMalformed )
////fmt.Println(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet)
//if Errors&jwt.ValidationErrorMalformed != 0 {
//	fmt.Println("That's not even a token")
//} else if Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
//	fmt.Println("Timing is everything")
//} else {
//	fmt.Println("Couldn't handle this token:")
//}

//a := ABC("456")
//fmt.Println(a.name())
//fmt.Println("4444444")
//theuuid := uuid.New()
//a := [16]byte(uuid.New())
//p ,_:=filepath.Abs("ssl/crt/meowalien_com.crt")
//fmt.Println("ssl/crt/meowalien_com.crt  --  ",p) // F:\GoTest\GoTest\master.exe <nil>
//fmt.Println(string(a[:]))
//fmt.Println(string(uuid.New().NodeID()))
//}
