package main

import (
	"github.com/AubSs/fasthttplogger"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"encoding/json"
	"flag"
	"fmt"
	"time"
)

const (
	ContentTypeStr = "text/plain; charset=utf8"
	TestResponse  = "Nothing here, but it's ok"
)

var db Database

func ScheduleEndpoint(ctx *fasthttp.RequestCtx) {
	var (
		ResultJson         []uint8
		strContentType     = []byte("Content-Type")
		strApplicationJSON = []byte("application/json")
	)

	Faculty := string(ctx.QueryArgs().Peek("faculty"))
	if Faculty == "" {
		ctx.Response.SetStatusCode(400)
		return
	}

	GroupID := string(ctx.QueryArgs().Peek("id"))
	if GroupID == "" {
		ctx.Response.SetStatusCode(400)
		return
	}

	InputDate := string(ctx.QueryArgs().Peek("date"))
	InputEndDate := string(ctx.QueryArgs().Peek("end_date"))

	if InputDate == "" && InputEndDate == "" {
		rasp, err := db.GetScheduleByDate(
			Faculty, GroupID,
			time.Now().Format("2006-01-02"),
		)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}

		if RawJson, err := json.Marshal(rasp); err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		} else {
			ResultJson = RawJson
		}
	} else if InputDate != "" && InputEndDate == "" {
		ParsedDate, err := time.Parse("2006-01-02", InputDate)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			return
		}

		ParsedDate = ParsedDate.In(time.Local)

		if rasp, err := db.GetScheduleByDate(
			Faculty, GroupID,
			ParsedDate.Format("2006-01-02"),
		); err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		} else {
			if RawJson, err := json.Marshal(rasp); err != nil {
				ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
			} else {
				ResultJson = RawJson
			}
		}
	} else if (InputDate != "" && InputEndDate != "") || (InputDate == "" && InputEndDate != "") {
		var Date time.Time

		if InputDate == "" {
			TimeNow := time.Now()
			Date = time.Date(
				TimeNow.Year(), TimeNow.Month(), TimeNow.Day(),
				0, 0, 0, 0,
				TimeNow.Location(),
			)
		} else {
			if ParsedDate, err := time.Parse("2006-01-02", InputDate); err != nil {
				ctx.Response.SetStatusCode(400)
				return
			} else {
				Date = ParsedDate
			}
		}

		ParsedEndDate, err := time.Parse("2006-01-02", InputEndDate)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			return
		}

		Date = Date.In(time.Local)
		ParsedEndDate = ParsedEndDate.In(time.Local)

		var buffer []primitive.M

		for Date := Date; Date.After(ParsedEndDate) == false; Date = Date.AddDate(0, 0, 1) {
			StringDate := Date.Format("2006-01-02")

			if DaySchedule, err := db.GetScheduleByDate(Faculty, GroupID, StringDate); err != nil {
				ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
			} else {
				if DaySchedule != nil {
					buffer = append(buffer, DaySchedule)
				}
			}
		}

		if RawJson, err := json.Marshal(buffer); err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		} else {
			ResultJson = RawJson
		}
	}

	ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
	ctx.Response.SetStatusCode(200)
	ctx.SetBody(ResultJson)
}

func PoolEndpoint(ctx *fasthttp.RequestCtx) {
	var (
		strContentType     = []byte("Content-Type")
		strApplicationJSON = []byte("application/json")
	)

	if tmp, err := db.Pool(); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	} else {
		json_obj, err := json.Marshal(FacultiesPool{len(tmp), tmp})
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}

		ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
		ctx.Response.SetStatusCode(200)
		ctx.SetBody(json_obj)
	}
}

func TestEndpoint(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType(ContentTypeStr)
	ctx.SetBodyString(TestResponse)
}

func main() {
	Port := flag.String("port", ":8080", "Port in format ':PORT'")
	Compress := flag.Bool("compress", false, "On/of compression")
	DbAddress := flag.String("db-address", "localhost", "MongoDB address")
	DbPort := flag.String("db-port", "27017", "MongoDB port")

	flag.Parse()

	if db_conn, err := InitDB(*DbAddress, *DbPort); err != nil {
		panic(fmt.Sprintf("%s", "Can't open connection to db, error - %s", err))
	} else {
		db = db_conn
	}
	defer db.Close()

	router := router.New()
	router.GET("/v1/test", TestEndpoint)
	router.GET("/v1/rasp", ScheduleEndpoint)
	router.GET("/v1/pool", PoolEndpoint)

	handler := fasthttplogger.Combined(router.Handler)
	if *Compress {
		handler = fasthttp.CompressHandler(handler)
	}

	if err := fasthttp.ListenAndServe(*Port, handler); err != nil {
		panic(fmt.Sprintf("Error in ListenAndServe: %s", err))
	}
}
