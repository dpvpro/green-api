package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
)

const (
    ApiSrv = "https://3100.api.green-api.com"
    LocalSrvPort = ":10001"
    FileName = "demo.file"
)

type PageData struct {
	IDInstance       string
	ApiTokenInstance string
	PhoneNumber      string
	Message          string
	URLFile          string
	Result           string
	Error            string
	// какой метод был вызван
	ApiMethod        string
}

var TmplSite *template.Template

func main() {
	TmplSite = template.Must(template.ParseFiles("index.tmpl"))

	// оработчик главной страницы
	http.HandleFunc("/", rootHandler)
	// обработчик css файла
	http.HandleFunc("/style.css", cssHandler)

	fmt.Printf("Сервер запущен на http://localhost%s\n", LocalSrvPort)
	log.Fatal(http.ListenAndServe(LocalSrvPort, nil))
}

func cssHandler (w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "style.css")
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// get без apiMethod — отдаём пустую форму
	if r.Method == http.MethodGet && r.URL.Query().Get("apiMethod") == "" {
		TmplSite.Execute(w, PageData{})
		return
	}

	// get или post — обработка формы
	r.ParseForm()

	data := PageData{
		IDInstance:       r.FormValue("idInstance"),
		ApiTokenInstance: r.FormValue("apiTokenInstance"),
		PhoneNumber:      r.FormValue("phoneNumber"),
		Message:          r.FormValue("message"),
		URLFile:          r.FormValue("urlFile"),
		ApiMethod:        r.FormValue("apiMethod"),
	}

	// вызываем нужный метод green api
	result, err := handlingForms(data)
	if err != nil {
		data.Error = err.Error()
	} else {
		data.Result = result
	}

	TmplSite.Execute(w, data)
}

func handlingForms(data PageData) (string, error) {
	var url string
	var method string
	var jsonBody []byte

	switch data.ApiMethod {
	case "getSettings":
		url = fmt.Sprintf("%s/waInstance%s/getSettings/%s", ApiSrv, data.IDInstance, data.ApiTokenInstance)
		method = http.MethodGet

	case "getStateInstance":
		url = fmt.Sprintf("%s/waInstance%s/getStateInstance/%s", ApiSrv, data.IDInstance, data.ApiTokenInstance)
		method = http.MethodGet

	case "sendMessage":
		url = fmt.Sprintf("%s/waInstance%s/sendMessage/%s", ApiSrv, data.IDInstance, data.ApiTokenInstance)
		method = http.MethodPost
		payload := map[string]string{
			"chatId":  data.PhoneNumber + "@c.us",
			"message": data.Message,
		}
		jsonBody, _ = json.Marshal(payload)

	case "sendFileByUrl":
		url = fmt.Sprintf("%s/waInstance%s/sendFileByUrl/%s", ApiSrv, data.IDInstance, data.ApiTokenInstance)
		method = http.MethodPost
		payload := map[string]string{
			"chatId":   data.PhoneNumber + "@c.us",
			"urlFile":  data.URLFile,
			"fileName": FileName,
		}
		jsonBody, _ = json.Marshal(payload)

	default:
		return "", fmt.Errorf("неизвестное действие: %s", data.ApiMethod)
	}

	return callGreenAPI(method, url, jsonBody)

}

func callGreenAPI(method, url string, bodyRaw []byte) (string, error) {

	var body io.Reader

	body = bytes.NewReader(bodyRaw)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	if bodyRaw != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	var prettyJson bytes.Buffer
	if err := json.Indent(&prettyJson, respBody, "", "  "); err != nil {
		return string(respBody), nil
	}

	return prettyJson.String(), nil
}
