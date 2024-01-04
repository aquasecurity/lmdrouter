package middleware

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/seantcanavan/lambda_jwt_router/internal/examples/books"
	"github.com/seantcanavan/lambda_jwt_router/lambda_jwt"
	"github.com/seantcanavan/lambda_jwt_router/lambda_router"
	"log"
	"net/http"
	"os"
)

var router *lambda_router.Router

func init() {
	// implement your own base middleware functions and add to the NewRouter declaration to apply to every route
	router = lambda_router.NewRouter("/api", lambda_jwt.InjectLambdaContextMW)

	// to configure middleware at the route level, add them singularly to each route
	router.Route("DELETE", "/books/:id", books.DeleteLambda, lambda_jwt.LogRequestMW)
	router.Route("GET", "/books/:id", books.GetLambda, lambda_jwt.LogRequestMW)
	router.Route("POST", "/books", books.CreateLambda, lambda_jwt.LogRequestMW)
	router.Route("PUT", "/books/:id", books.UpdateLambda, lambda_jwt.LogRequestMW)
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
