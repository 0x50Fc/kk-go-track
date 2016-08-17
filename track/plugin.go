package track

import (
	"database/sql"
	"github.com/hailongz/kk-go-db/kk"
	"github.com/hailongz/kk-go-task/task"
)

type Track struct {
	Id   int64 `json:"id"`
	Code int64 `json:"code"`

	IP          string  `json:"ip"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	TimeZone    string  `json:"timeZone"`
	Continent   string  `json:"continent"`
	CountryCode string  `json:"countryCode"`
	Country     string  `json:"country"`
	Province    string  `json:"province"`
	City        string  `json:"city"`
	PostalCode  string  `json:"postalCode"`

	Mtime int64 `json:"mtime"` //修改时间
	Ctime int64 `json:"ctime"` //创建时间
}

var TrackTable = kk.DBTable{"track",

	"id",

	map[string]kk.DBField{"code": kk.DBField{0, kk.DBFieldTypeInt64},
		"ip":          kk.DBField{32, kk.DBFieldTypeString},
		"latitude":    kk.DBField{0, kk.DBFieldTypeDouble},
		"longitude":   kk.DBField{0, kk.DBFieldTypeDouble},
		"timezone":    kk.DBField{32, kk.DBFieldTypeString},
		"continent":   kk.DBField{32, kk.DBFieldTypeString},
		"countrycode": kk.DBField{4, kk.DBFieldTypeString},
		"country":     kk.DBField{64, kk.DBFieldTypeString},
		"province":    kk.DBField{128, kk.DBFieldTypeString},
		"city":        kk.DBField{128, kk.DBFieldTypeString},
		"postalcode":  kk.DBField{64, kk.DBFieldTypeString},
		"mtime":       kk.DBField{0, kk.DBFieldTypeInt},
		"ctime":       kk.DBField{0, kk.DBFieldTypeInt}},

	map[string]kk.DBIndex{"code": kk.DBIndex{"code", kk.DBIndexTypeAsc, true}}}

type Plugin struct {
	Db     *sql.DB
	Prefix string
	Geodb  string
}

func Load(context *task.Context) error {

	var db = context.Get("db").(*sql.DB)
	var p = Plugin{db, context.Get("prefix").(string), context.Get("geodb").(string)}

	var err = kk.DBBuild(db, &TrackTable, p.Prefix, 1)

	if err != nil {
		return err
	}

	context.Plugin(&p)(&TrackService{})(&TrackTask{}, &TrackSetTask{}, &TrackLoadTask{}, &TrackUnLoadTask{})

	return nil
}
