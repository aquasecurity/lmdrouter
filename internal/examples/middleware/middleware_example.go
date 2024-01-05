package middleware

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/seantcanavan/lambda_jwt_router/internal/examples/books"
	"github.com/seantcanavan/lambda_jwt_router/lmw"
	"github.com/seantcanavan/lambda_jwt_router/lrtr"
	"log"
	"net/http"
	"os"
)

var router *lrtr.Router

func init() {
	// implement your own base middleware functions and add to the NewRouter declaration to apply to every route
	router = lrtr.NewRouter("/api", lmw.InjectLambdaContextMW)

	// to configure middleware at the route level, add them singularly to each route
	router.Route("DELETE", "/books/:id", books.DeleteLambda, lmw.LogRequestMW)
	router.Route("GET", "/books/:id", books.GetLambda, lmw.LogRequestMW)
	router.Route("POST", "/books", books.CreateLambda, lmw.LogRequestMW)
	router.Route("PUT", "/books/:id", books.UpdateLambda, lmw.LogRequestMW)
}

func main() {
	// if we're running this in staging or production, we want to use the lambda handler on startup
	environment := os.Getenv("STAGE")
	if environment == "staging" || environment == "production" {
		lambda.Start(router.Handler)
	} else { // else, we want to start an HTTP server to listen for local development
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("Ready to listen and serve on port %s", port)
		err := http.ListenAndServe(":"+port, http.HandlerFunc(router.ServeHTTP))
		if err != nil {
			panic(fmt.Sprintf("http.ListAndServe error %s", err))
		}
	}
}
