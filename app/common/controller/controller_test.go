package common_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krishamoud/game/app/common/controller"
	. "github.com/smartystreets/goconvey/convey"
)

func TestControllerSpec(t *testing.T) {
	Convey("Given a running server and controller instance", t, func() {
		c := common.Controller{}
		mux := http.NewServeMux()
		server := httptest.NewServer(mux)
		defer server.Close()
		Convey("When SendJSON is called from handler with 200", func() {
			mux.HandleFunc("/test1", func(w http.ResponseWriter, r *http.Request) {
				v := struct {
					Message string `json:"message"`
				}{
					Message: "Hello",
				}
				c.SendJSON(w, nil, &v, http.StatusOK)
			})
			resp, err := http.Get(server.URL + "/test1")
			Convey("Then response should be 200 with correct JSON", func() {
				if err != nil {
					t.Fatal(err)
				}
				body, err := ioutil.ReadAll(resp.Body)
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusOK)
				So(string(body), ShouldEqual, `{"message":"Hello"}`)
			})
		})
		Convey("When SendJSON is called from handler with 400", func() {
			mux.HandleFunc("/test2", func(w http.ResponseWriter, r *http.Request) {
				c.SendJSON(w, nil, nil, http.StatusBadRequest)
			})
			resp, err := http.Get(server.URL + "/test2")
			Convey("Then response should be 400 with correct JSON", func() {
				So(err, ShouldBeNil)
				So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
			})
		})
	})
}
