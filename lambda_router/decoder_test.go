package lambda_router

import (
	"cloud.google.com/go/civil"
	"github.com/seantcanavan/lambda_jwt_router/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

type mockConst string

type Number string

const (
	numberOne   Number = "one"
	numberTwo   Number = "two"
	numberThree Number = "three"
)

type mockItem struct {
	ID   string
	Name string
	Date time.Time
}

type mockGetReq struct {
	ID            string `lambda:"path.id"`
	ShowSomething bool   `lambda:"query.show_something"`
}

type mockListReq struct {
	Alias         stringAliasExample    `lambda:"query.alias"`
	AliasPtr      *stringAliasExample   `lambda:"query.alias_ptr"`
	Bool1         bool                  `lambda:"query.bool1"`
	Bool2         bool                  `lambda:"query.bool2"`
	Bool3         bool                  `lambda:"query.bool3"`
	Bool4         bool                  `lambda:"query.bool4"`
	Bool5         bool                  `lambda:"query.bool5"`
	Bool6         bool                  `lambda:"query.bool6"`
	Bool7         bool                  `lambda:"query.bool7"`
	Bool8         bool                  `lambda:"query.bool8"`
	Bool9         bool                  `lambda:"query.bool9"`
	Civil         civil.Date            `lambda:"query.civil"`
	CivilPtr      *civil.Date           `lambda:"query.civilPtr"`
	CivilPtrNil   *civil.Date           `lambda:"query.civilPtrNil"`
	CommaSplit    []Number              `lambda:"query.commaSplit"`
	CommaSplitPtr []*Number             `lambda:"query.commaSplitPtr"`
	Const         mockConst             `lambda:"query.const"`
	ConstPtr      *mockConst            `lambda:"query.constPtr"`
	ConstPtrNil   *mockConst            `lambda:"query.constPtrNil"`
	Encoding      []string              `lambda:"header.Accept-Encoding"`
	ID            string                `lambda:"path.id"`
	IDs           []*string             `lambda:"query.ids"`
	Language      string                `lambda:"header.Accept-Language"`
	MongoID       primitive.ObjectID    `lambda:"query.mongoId"`
	MongoIDPtr    *primitive.ObjectID   `lambda:"query.mongoIdPtr"`
	MongoIDPtrNil *primitive.ObjectID   `lambda:"query.mongoIdPtrNil"`
	MongoIDs      []primitive.ObjectID  `lambda:"query.mongoIds"`
	MongoIDsPtr   []*primitive.ObjectID `lambda:"query.mongoIdsPtr"`
	Number        *float32              `lambda:"query.number"`
	Numbers       []float64             `lambda:"query.numbers"`
	PBoolOne      *bool                 `lambda:"query.pbool1"`
	PBoolTwo      *bool                 `lambda:"query.pbool2"`
	Page          int64                 `lambda:"query.page"`
	PageSize      *int64                `lambda:"query.page_size"`
	Terms         []string              `lambda:"query.terms"`
	Time          time.Time             `lambda:"query.time"`
	TimePtr       *time.Time            `lambda:"query.timePtr"`
	TimePtrNil    *time.Time            `lambda:"query.timePtrNil"`
}

type mockPostReq struct {
	ID   string    `lambda:"path.id"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type stringAliasExample string

const aliasExample stringAliasExample = "world"

func TestMarshalLambdaRequest(t *testing.T) {
	mi := mockItem{
		ID:   util.GenerateRandomString(10),
		Name: util.GenerateRandomString(10),
	}

	t.Run("verify MarshalReq correctly adds the JSON string to the request body", func(t *testing.T) {
		req := MarshalReq(mi)

		var miParsed mockItem
		err := UnmarshalReq(req, true, &miParsed)
		assert.Nil(t, err)
		require.Equal(t, mi.ID, miParsed.ID)
		require.Equal(t, mi.Name, miParsed.Name)
	})
}

func Test_UnmarshalReq(t *testing.T) {
	t.Run("valid path&query input", func(t *testing.T) {
		mongoID1 := primitive.NewObjectID()
		mongoID2 := primitive.NewObjectID()

		var input mockListReq
		err := UnmarshalReq(
			events.APIGatewayProxyRequest{
				PathParameters: map[string]string{
					"id": "fake-scan-id",
				},
				QueryStringParameters: map[string]string{
					"alias":         "hello",
					"alias_ptr":     "world",
					"bool1":         "1",
					"bool2":         "true",
					"bool3":         "on",
					"bool4":         "enabled",
					"bool5":         "t",
					"bool6":         "TRUE",
					"bool7":         "ON",
					"bool8":         "ENABLED",
					"bool9":         "T",
					"civil":         "2023-12-22",
					"civilPtr":      "2024-12-22",
					"commaSplit":    "one,two,three",
					"commaSplitPtr": "one,two,three",
					"const":         "twenty",
					"constPtr":      "thirty",
					"mongoId":       mongoID1.Hex(),
					"mongoIdPtr":    mongoID1.Hex(),
					"number":        "90.10982",
					"page":          "2",
					"page_size":     "30",
					"pbool1":        "0",
					"time":          "2021-11-01T11:11:11.000Z",
					"timePtr":       "2021-11-01T11:11:11.000Z",
				},
				MultiValueQueryStringParameters: map[string][]string{
					"commaSplits": {"four,five,six"},
					"ids":         {"7", "8", "9"},
					"mongoIds":    {mongoID1.Hex(), mongoID2.Hex()},
					"mongoIdsPtr": {mongoID1.Hex(), mongoID2.Hex()},
					"numbers":     {"1.2", "3.5", "666.666"},
					"terms":       {"artist", "label"},
				},
				Headers: map[string]string{
					"Accept-Language": "en-us",
				},
				MultiValueHeaders: map[string][]string{
					"Accept-Encoding": {"gzip", "deflate"},
				},
			},
			false,
			&input,
		)
		require.NoError(t, err)

		require.Equal(t, *input.AliasPtr, stringAliasExample("world"))
		require.Equal(t, input.Alias, stringAliasExample("hello"))
		require.Equal(t, input.Bool1, true)
		require.Equal(t, input.Bool2, true)
		require.Equal(t, input.Bool3, true)
		require.Equal(t, input.Bool4, true)
		require.Equal(t, input.Bool5, true)
		require.Equal(t, input.Bool6, true)
		require.Equal(t, input.Bool7, true)
		require.Equal(t, input.Bool8, true)
		require.Equal(t, input.Bool9, true)
		require.Equal(t, input.Civil.String(), "2023-12-22")
		require.Equal(t, input.CivilPtr.String(), "2024-12-22")
		require.Equal(t, input.Const, mockConst("twenty"))
		require.Equal(t, *input.ConstPtr, mockConst("thirty"))
		require.Equal(t, input.ID, "fake-scan-id")
		require.Equal(t, input.Language, "en-us")
		require.Equal(t, input.MongoID, mongoID1)
		require.Equal(t, input.MongoIDPtr.Hex(), mongoID1.Hex())
		require.Equal(t, input.Number, func() *float32 { a := float32(90.10982); return &a }())
		require.Equal(t, *input.PBoolOne, false)
		require.Equal(t, input.Page, int64(2))
		require.Equal(t, input.PageSize, func() *int64 { a := int64(30); return &a }())
		require.Equal(t, input.Time, time.Date(2021, 11, 1, 11, 11, 11, 0, time.UTC))
		require.Equal(t, *input.TimePtr, time.Date(2021, 11, 1, 11, 11, 11, 0, time.UTC))

		numberPtrs := []*Number{
			func() *Number { a := Number("one"); return &a }(),
			func() *Number { a := Number("two"); return &a }(),
			func() *Number { a := Number("three"); return &a }(),
		}

		idPtrs := []*string{
			func() *string { a := "7"; return &a }(),
			func() *string { a := "8"; return &a }(),
			func() *string { a := "9"; return &a }(),
		}

		require.EqualValues(t, input.CommaSplit, []Number{"one", "two", "three"})
		require.EqualValues(t, input.CommaSplitPtr, numberPtrs)
		require.EqualValues(t, input.Encoding, []string{"gzip", "deflate"})
		require.EqualValues(t, input.IDs, idPtrs)
		require.EqualValues(t, input.MongoIDs, []primitive.ObjectID{mongoID1, mongoID2})
		require.EqualValues(t, input.MongoIDsPtr, []*primitive.ObjectID{&mongoID1, &mongoID2})
		require.EqualValues(t, input.Numbers, []float64{1.2, 3.5, 666.666})
		require.EqualValues(t, input.Terms, []string{"artist", "label"})

		require.Nil(t, input.CivilPtrNil)
		require.Nil(t, input.ConstPtrNil)
		require.Nil(t, input.MongoIDPtrNil)
		require.Nil(t, input.PBoolTwo)
		require.Nil(t, input.TimePtrNil)
	})
	t.Run("valid empty input", func(t *testing.T) {
		var input mockListReq
		err := UnmarshalReq(
			events.APIGatewayProxyRequest{},
			false,
			&input,
		)
		require.NoError(t, err)
	})
	t.Run("valid input unset values", func(t *testing.T) {
		var input mockListReq
		err := UnmarshalReq(
			events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"mongoId": "",
				},
			},
			false,
			&input,
		)
		require.NoError(t, err)
	})

	t.Run("invalid path&query input", func(t *testing.T) {
		var input mockListReq
		err := UnmarshalReq(
			events.APIGatewayProxyRequest{
				PathParameters: map[string]string{
					"id": "fake-scan-id",
				},
				QueryStringParameters: map[string]string{
					"page": "abcd",
				},
			},
			false,
			&input,
		)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), "page"))
		require.True(t, strings.Contains(err.Error(), "must be a valid integer"))
	})

	fakeDate := time.Date(2020, 3, 23, 11, 33, 0, 0, time.UTC)

	t.Run("valid body input, not base64", func(t *testing.T) {
		var input mockPostReq
		err := UnmarshalReq(
			events.APIGatewayProxyRequest{
				IsBase64Encoded: false,
				PathParameters: map[string]string{
					"id": "bla",
				},
				Body: `{"name":"Fake Post","date":"2020-03-23T11:33:00Z"}`,
			},
			true,
			&input,
		)

		require.Equal(t, nil, err, "ErrorRes must be nil")
		require.Equal(t, "bla", input.ID, "ID must be parsed from path parameters")
		require.Equal(t, "Fake Post", input.Name, "Name must be parsed from body")
		require.Equal(t, fakeDate, input.Date, "Date must be parsed from body")
	})

	t.Run("invalid body input, not base64", func(t *testing.T) {
		var input mockPostReq
		err := UnmarshalReq(
			events.APIGatewayProxyRequest{
				IsBase64Encoded: false,
				Body:            `this is not JSON`,
			},
			true,
			&input,
		)

		assert.NotEqual(t, nil, err, "ErrorRes must not be nil")
	})

	t.Run("valid body input, base64", func(t *testing.T) {
		var input mockPostReq
		err := UnmarshalReq(
			events.APIGatewayProxyRequest{
				IsBase64Encoded: true,
				Body:            "eyJuYW1lIjoiRmFrZSBQb3N0IiwiZGF0ZSI6IjIwMjAtMDMtMjNUMTE6MzM6MDBaIn0=",
			},
			true,
			&input,
		)

		require.Equal(t, nil, err, "ErrorRes must be nil")
		require.Equal(t, "Fake Post", input.Name, "Name must be parsed from body")
		require.Equal(t, fakeDate, input.Date, "Date must be parsed from body")
	})

	t.Run("invalid body input, base64", func(t *testing.T) {
		var input mockPostReq
		err := UnmarshalReq(
			events.APIGatewayProxyRequest{
				IsBase64Encoded: true,
				Body:            "dGhpcyBpcyBub3QgSlNPTg==",
			},
			true,
			&input,
		)

		assert.NotEqual(t, nil, err, "ErrorRes must not be nil")
	})
}
