package track

import (
	"database/sql"
	kkdb "github.com/hailongz/kk-go-db/kk"
	"github.com/hailongz/kk-go-task/task"
	"github.com/hailongz/kk-go/kk"
	geoip2 "github.com/oschwald/geoip2-golang"
	"log"
	"net"
	"time"
)

const TRACK_EXPIRES_SECOND = 6

type TrackCache struct {
	track   *Track
	expires int64
}

type TrackService struct {
	task.Service
	dispatch *kk.Dispatch
	tracks   map[int64]TrackCache
	geodb    *geoip2.Reader
}

func (S *TrackService) Handle(task task.ITask) error {
	return S.ReflectHandle(task, S)
}

/**
 * 启动追踪
 */
func (S *TrackService) HandleTrackLoadTask(task *TrackLoadTask) error {

	var plugin = S.Plugin().(*Plugin)

	S.tracks = make(map[int64]TrackCache)

	if S.dispatch == nil {

		S.dispatch = kk.NewDispatch()

		var cleanup func() = nil

		cleanup = func() {

			var keys []int64
			var now = time.Now().Unix()

			for key, v := range S.tracks {
				if now > v.expires {
					keys = append(keys, key)
				}
			}

			for _, key := range keys {
				delete(S.tracks, key)
			}

			S.dispatch.AsyncDelay(cleanup, TRACK_EXPIRES_SECOND*time.Second)
		}

		S.dispatch.AsyncDelay(cleanup, TRACK_EXPIRES_SECOND*time.Second)

	}

	if S.geodb == nil {

		db, err := geoip2.Open(plugin.Geodb)

		if err != nil {
			log.Fatal(err)
		}

		S.geodb = db

	}

	return nil
}

/**
 * 卸载追踪
 */
func (S *TrackService) HandleTrackUnLoadTask(task *TrackLoadTask) error {

	if S.dispatch != nil {
		S.dispatch.Break()
		S.dispatch = nil
	}

	if S.geodb != nil {
		S.geodb.Close()
		S.geodb = nil
	}

	S.tracks = nil

	return nil
}

func (S *TrackService) SetTrack(track *Track, db *sql.DB, prefix string) {

	go func() {

		var v *Track = nil

		S.dispatch.Sync(func() {
			var vv, ok = S.tracks[track.Code]
			if ok {
				v = vv.track
			}
		})

		if v == nil {
			v = new(Track)
			v.Code = track.Code
			v.Ctime = time.Now().Unix()
		}

		if track.IP != "" {
			v.IP = track.IP
			ip := net.ParseIP(v.IP)
			city, err := S.geodb.City(ip)
			if err != nil {
				log.Fatal(err)
			} else {
				v.Continent = city.Continent.Names["en"]
				v.Country = city.Country.Names["en"]
				v.CountryCode = city.Country.IsoCode
				if len(city.Subdivisions) > 0 {
					v.Province = city.Subdivisions[0].Names["en"]
				}
				v.City = city.City.Names["en"]
				v.PostalCode = city.Postal.Code
				v.TimeZone = city.Location.TimeZone
				v.Latitude = city.Location.Latitude
				v.Longitude = city.Location.Longitude
			}
		}

		if track.Latitude != 0 && track.Longitude != 0 {
			v.Latitude = track.Latitude
			v.Longitude = track.Longitude
		}

		v.Mtime = time.Now().Unix()

		if v.Id == 0 {
			var _, err = kkdb.DBInsert(db, &TrackTable, prefix, v)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			var _, err = kkdb.DBUpdate(db, &TrackTable, prefix, v)
			if err != nil {
				log.Fatal(err)
			}
		}

		S.dispatch.Async(func() {
			S.tracks[v.Code] = TrackCache{v, time.Now().Unix() + TRACK_EXPIRES_SECOND}
		})
	}()

}

/**
 * 更新追踪
 */
func (S *TrackService) HandleTrackSetTask(task *TrackSetTask) error {

	var plugin = S.Plugin().(*Plugin)
	var db = plugin.Db

	if task.Code == 0 {
		if task.IP == "" {
			task.Result.Errno = ERRNO_NOT_FOUND_IP
			task.Result.Errmsg = "未找到IP地址"
			return nil
		}
		task.Code = kk.UUID()
	}

	task.Result.Code = task.Code

	var v = Track{}
	v.Code = task.Code
	v.IP = task.IP
	v.Latitude = task.Latitude
	v.Longitude = task.Longitude
	S.SetTrack(&v, db, plugin.Prefix)

	log.Println("TrackService.HandleTrackSetTask")

	return nil

}

/**
 * 获取追踪
 */
func (S *TrackService) HandleTrackTask(task *TrackTask) error {

	if task.Code == 0 {
		task.Result.Errno = ERRNO_NOT_FOUND_CODE
		task.Result.Errmsg = "未找到跟踪码"
	}

	var v *Track = nil

	S.dispatch.Sync(func() {
		var vv, ok = S.tracks[task.Code]
		if ok {
			v = vv.track
			vv.expires = time.Now().Unix() + TRACK_EXPIRES_SECOND
		}
	})

	if v == nil {
		var plugin = S.Plugin().(*Plugin)
		var db = plugin.Db
		var rs, err = kkdb.DBQuery(db, &TrackTable, plugin.Prefix, " WHERE code=?", task.Code)

		if err != nil {
			task.Result.Errno = ERRNO_TRACK
			task.Result.Errmsg = err.Error()
			return nil
		}

		defer rs.Close()

		var vv = Track{}
		var scaner = kkdb.NewDBScaner(&vv)

		if rs.Next() {
			err = scaner.Scan(rs)
			if err == nil {
				v = &vv
				S.dispatch.Async(func() {
					S.tracks[v.Code] = TrackCache{v, time.Now().Unix() + TRACK_EXPIRES_SECOND}
				})
			} else {
				task.Result.Errno = ERRNO_TRACK
				task.Result.Errmsg = err.Error()
				return nil
			}
		}
	}

	task.Result.Track = v

	return nil
}
