package main

import (
	"github.com/AubSs/fasthttplogger"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"bytes"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"time"
)

var (
	db  Database
	cfg Configuration
)

func groups_endpoint(ctx *fasthttp.RequestCtx) {
	var (
		json_obj           []uint8
		err                error
		strContentType     = []byte("Content-Type")
		strApplicationJSON = []byte("application/json")
	)

	auth := ctx.Request.Header.Peek("Authorization")

	if resp, err := db.CheckToken(auth); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	} else if resp != 1 {
		ctx.Response.SetStatusCode(401)
		return
	}
	
	group_id := string(ctx.QueryArgs().Peek("id"))
	if group_id == "" {
		ctx.Response.SetStatusCode(400)
		return
	}

	rasp, err := db.Search(group_id)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
	
	date := string(ctx.QueryArgs().Peek("date"))
	end_date := string(ctx.QueryArgs().Peek("end_date"))

	if date == "" && end_date == "" {
		json_obj, err = json.Marshal(rasp)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}
	} else if date != "" && end_date == "" {
		date_parsed, err := time.Parse("2006-01-02", date)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			return
		}

		date_parsed = date_parsed.In(time.Local)

		tmp := rasp[date_parsed.Format("2006-01-02")]
		if tmp != nil {
			json_obj, err = json.Marshal(tmp)
			if err != nil {
				ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
			}
		} else {
			ctx.Response.SetStatusCode(404)
			return
		}
	} else if date != "" && end_date != "" {
		date_parsed, err := time.Parse("2006-01-02", date)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			return
		}

		end_parsed, err := time.Parse("2006-01-02", end_date)
		if err != nil {
			ctx.Response.SetStatusCode(400)
			return
		}

		date_parsed = date_parsed.In(time.Local)
		end_parsed = end_parsed.In(time.Local)

		res := make(map[string]primitive.M)

		for date_parsed := date_parsed; date_parsed.After(end_parsed) == false; date_parsed = date_parsed.AddDate(0, 0, 1) {
			d_now := date_parsed.Format("2006-01-02")
			tmp := rasp[d_now]

			if tmp != nil {
				res[d_now] = tmp.(primitive.M)
			}
		}

		json_obj, err = json.Marshal(res)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}
	}

	ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
	ctx.Response.SetStatusCode(200)
	fmt.Fprintf(ctx, string(json_obj))
	return
}

func groups_pool_endpoint(ctx *fasthttp.RequestCtx) {
	var (
		strContentType     = []byte("Content-Type")
		strApplicationJSON = []byte("application/json")
	)

	auth := ctx.Request.Header.Peek("Authorization")
	if resp, err :=db.CheckToken(auth); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	} else if resp != 1 {
		ctx.Response.SetStatusCode(401)
		return
	}
	
	tmp, err := db.Pool()
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}

	json_obj, err := json.Marshal(Pool{Count: len(tmp), Payload: tmp})
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}

	ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
	ctx.Response.SetStatusCode(200)
	fmt.Fprintf(ctx, string(json_obj))
}

func token_endpoint(ctx *fasthttp.RequestCtx) {
	var (
		strContentType     = []byte("Content-Type")
		strApplicationJSON = []byte("application/json")
	)

	auth := ctx.Request.Header.Peek("Authorization")
	if bytes.Compare(auth, cfg.Password) != 0{
		ctx.Response.SetStatusCode(401)
		return
	} else {
		token := make([]byte, 10)
		
		for {			
			_, err := rand.Read(token)
			if err != nil {
				ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
			}

			db_response, err := db.InsertToken(token)
			if err != nil {
				ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
				break
			}
			
			if db_response == 1 {
				break
			}
		}

		resp, err := json.Marshal(Token{token})
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}

		ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
		ctx.Response.SetStatusCode(200)
		fmt.Fprintf(ctx, string(resp))
	}
}

func token_check_endpoint(ctx *fasthttp.RequestCtx) {
	auth := ctx.Request.Header.Peek("Authorization")
	
	if resp, err := db.CheckToken(auth); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	} else if resp == 1 {
		ctx.Response.SetStatusCode(200)
	} else if resp == -1 {
		ctx.Response.SetStatusCode(401)
	}
}

func test_endpoint(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/plain; charset=utf8")
	fmt.Fprintf(ctx, "OK!")
}

func main() {
	password_for_token := flag.String("passwd", "", "password for token generation")
	api_port := flag.String("port", ":8080", "Port in format ':PORT'")
	api_compress := flag.Bool("compress", false, "On/of compression")
	db_address := flag.String("db-address", "localhost", "MongoDB address")
	db_port := flag.String("db-port", "27017", "MongoDB port")

	flag.Parse()

	cfg = Configuration{
		Password:   []byte(*password_for_token),
		Port:       *api_port,
		Compress:   *api_compress,
		DB_address: *db_address,
		DB_port:    *db_port,
	}

	if db_conn, err := db_init(cfg.DB_address, cfg.DB_port); err != nil {
		panic(fmt.Sprintf("%s", "Can't open connection to db, error - %s", err))
	} else {
		db = db_conn
	}
	
	router := router.New()
	router.GET("/test", test_endpoint)
	router.GET("/token", token_endpoint)
	router.GET("/token/check", token_check_endpoint)
	router.GET("/rasp/groups", groups_endpoint)
	router.GET("/pool/groups", groups_pool_endpoint)

	handler := fasthttplogger.Combined(router.Handler)
	if cfg.Compress {
		handler = fasthttp.CompressHandler(handler)
	}

	if err := fasthttp.ListenAndServe(cfg.Port, handler); err != nil {
		if err := db.Close(); err != nil {
			fmt.Printf("%s", "Can't close connection to db, error - %s", err)
		}
		
		panic(fmt.Sprintf("Error in ListenAndServe: %s", err))
	}

	if err := db.Close(); err != nil {
		fmt.Printf("%s", "Can't close connection to db, error - %s", err)
	}
}
