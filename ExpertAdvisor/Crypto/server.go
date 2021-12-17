package expert

import (
	//"golang.org/x/net/websocket"
	"net/http"
)

// GET /user
// GET /user/{id}
// POST /users

type Server struct {
	Expert 		*ExpertAdvisorCrypto
	Client 		*http.Client
	IP			string
	Prices		map[string]*RenkoSerie
}

func NewRestAPI(expert *ExpertAdvisorCrypto,port string) *Server {
	rest := &Server{
		Expert: expert,
		Client: &http.Client{},
		IP: "localhost:"+port,
		Prices: make(map[string]*RenkoSerie),
	}
	return rest
}

func (s *Server) Run() {
	http.HandleFunc("/subscribe",Handler)
	http.HandleFunc("/unsubscribe",Handler)

	http.ListenAndServe(s.IP,nil)
}

func Handler(w http.ResponseWriter,r *http.Request) {

}

/*
httpposturl := "localhost:8080"
fmt.Println("HTTP JSON POST URL:", httpposturl)

var jsonData = []byte(`{
	"name": "morpheus",
	"job": "leader"
}`)
request, error := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonData))
request.Header.Set("Content-Type", "application/json; charset=UTF-8")

client := &http.Client{}
response, error := client.Do(request)
if error != nil {
	panic(error)
}
defer response.Body.Close()

fmt.Println("response Status:", response.Status)
fmt.Println("response Headers:", response.Header)
body, _ := ioutil.ReadAll(response.Body)
fmt.Println("response Body:", string(body))
*/