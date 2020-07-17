package main

import (
       "github.com/valyala/fasthttp"
       "go.mongodb.org/mongo-driver/bson"
       "go.mongodb.org/mongo-driver/bson/primitive"
       "github.com/AubSs/fasthttplogger"

       "flag"
       "encoding/json"
       "crypto/rand"
       "fmt"
       "time"
)
     
var (
    db Database
    cfg Configuration
)

func preps_endpoint(ctx *fasthttp.RequestCtx) {
     var (
     	 json_obj []uint8
	 err error
     	 strContentType = []byte("Content-Type")
     	 strApplicationJSON = []byte("application/json")
     )

     auth := string(ctx.Request.Header.Peek("Authorization"))
     if db.check_token(auth) != 1 {
     	ctx.Response.SetStatusCode(401)
	return
     } else {
	prep_name := string(ctx.QueryArgs().Peek("name"))
	if prep_name == "" {
	   ctx.Response.SetStatusCode(400)
	   return
	}

	rasp := db.search("preps", bson.M{"prep" : bson.M{"$regex" : prep_name}})

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

	     res := make(map[string] primitive.M)

	     for date_parsed := date_parsed; date_parsed.After(end_parsed) == false; date_parsed = date_parsed.AddDate(0,0,1) {
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
}

func groups_endpoint(ctx *fasthttp.RequestCtx) {
     var (
     	 json_obj []uint8
     	 err error
     	 strContentType = []byte("Content-Type")
     	 strApplicationJSON = []byte("application/json")
     )

     auth := string(ctx.Request.Header.Peek("Authorization"))
     if db.check_token(auth) != 1 {
     	ctx.Response.SetStatusCode(401)
	return
     } else {
	group_id := string(ctx.QueryArgs().Peek("id"))
	if group_id == "" {
	   ctx.Response.SetStatusCode(400)
	   return
	}

	rasp := db.search("groups", bson.M{"group" : group_id})

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

	     res := make(map[string] primitive.M)

	     for date_parsed := date_parsed; date_parsed.After(end_parsed) == false; date_parsed = date_parsed.AddDate(0,0,1) {
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
}

func groups_pool_endpoint(ctx *fasthttp.RequestCtx) {
     var (
     	 strContentType = []byte("Content-Type")
     	 strApplicationJSON = []byte("application/json")
     )

     auth := string(ctx.Request.Header.Peek("Authorization"))
     if db.check_token(auth) != 1 {
     	ctx.Response.SetStatusCode(401)
	return
     } else {
	tmp := db.pool("groups")

	json_obj, err := json.Marshal(Pool{Count: len(tmp), Payload: tmp})
	if err != nil {
     	   ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
     	}

	ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
     	ctx.Response.SetStatusCode(200)
	fmt.Fprintf(ctx, string(json_obj))
     }
}

func prep_pool_endpoint(ctx *fasthttp.RequestCtx) {
     var (
     	 strContentType = []byte("Content-Type")
     	 strApplicationJSON = []byte("application/json")
     )

     auth := string(ctx.Request.Header.Peek("Authorization"))
     if db.check_token(auth) != 1 {
     	ctx.Response.SetStatusCode(401)
	return
     } else {
	tmp := db.pool("preps")

	json_obj, err := json.Marshal(Pool{Count: len(tmp), Payload: tmp})
	if err != nil {
     	   ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
     	}

	ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
     	ctx.Response.SetStatusCode(200)
	fmt.Fprintf(ctx, string(json_obj))
     }
}

func token_endpoint(ctx *fasthttp.RequestCtx) {
     var (
     	 token string
     	 strContentType = []byte("Content-Type")
     	 strApplicationJSON = []byte("application/json")
     )

     auth := string(ctx.Request.Header.Peek("Authorization"))
     if auth != cfg.Password {
     	ctx.Response.SetStatusCode(401)
	return
     } else {
        for {
       	    tmp := make([]byte, 10)

     	    _, err := rand.Read(tmp)
     	    if err != nil {
     	       ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
     	    }

     	    token = fmt.Sprintf("%x", tmp)

            db_Response := db.insert_token(token)
            if db_Response == 1 {
                break
            }
     	}

     	resp, err := json.Marshal(Token{Key: token})
     	if err != nil {
     	   ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
     	}

     	ctx.Response.Header.SetCanonical(strContentType, strApplicationJSON)
     	ctx.Response.SetStatusCode(200)
     	fmt.Fprintf(ctx, string(resp))
     }
}

func test_endpoint(ctx *fasthttp.RequestCtx) {
     ctx.SetContentType("text/plain; charset=utf8")
     fmt.Fprintf(ctx, "OK!")
}

func empty_endpoint(ctx *fasthttp.RequestCtx) {
     ctx.Response.SetStatusCode(404)
     return
}

func main() {
     password_for_token := flag.String("passwd", "", "password for token generation")
     api_port := flag.String("port", ":8080", "Port in format ':PORT'")
     api_compress := flag.Bool("compress", false, "On/of compression")
     db_address := flag.String("db-address", "localhost", "MongoDB address")
     db_port := flag.String("db-port", "27017", "MongoDB port")

     flag.Parse()

     cfg = Configuration{
		Password : *password_for_token,
		Port : *api_port,
		Compress : *api_compress,
		DB_address : *db_address,
		DB_port: *db_port,
	    }
	    
     db = db_init(cfg.DB_address, cfg.DB_port)

     handler := func(ctx *fasthttp.RequestCtx) {
     	     switch string(ctx.Path()) {
	     	    case "/token":
		    	token_endpoint(ctx)
		    case "/rasp/group":
		    	 groups_endpoint(ctx)
		    case "/rasp/prep":
		    	 preps_endpoint(ctx)
		    case "/rasp/prep/pool":
		    	prep_pool_endpoint(ctx)
		    case "/rasp/group/pool":
		    	groups_pool_endpoint(ctx)
		    case "/test":
			test_endpoint(ctx)
		    default:
			empty_endpoint(ctx)
	     }
     }
     
     if cfg.Compress {
	handler = fasthttp.CompressHandler(handler)
     }

     if err := fasthttp.ListenAndServe(cfg.Port, fasthttplogger.Combined(handler)); err != nil {
	panic(fmt.Sprintf("Error in ListenAndServe: %s", err))
     }    
}