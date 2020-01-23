/*A continuación les mando adjunto un programa de como obtener información
de https://api.estadisticasbcra.com, terminando el ejempo de ayer. Les voy a pedir
que se registren en https://estadisticasbcra.com/api/registracion y obtengan un bearer
token para poder autenticarse individualmente (usenlo bien recuerden que tienen 100 requests por día).
Quiero que hagan un programa que para una fecha determinada que yo ingrese por consola me diga la
cotización del dólar oficial, la del blue y calcule cuál fue porcentaje de variación.
Por otro lado, para el rango de fechas 28-10-2019 hasta hoy, quiero
que busquen cuál fue el mejor día para hacer puré, es decir, comprar dólar oficial,
venderlo al blue y quedarme con una diferencia. A su vez, quiero saber cuál fue el
mejor día para comprar dólar blue.
*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const (
	authToken = "BEARER eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTA3Mzc2ODksInR5cGUiOiJleHRlcm5hbCIsInVzZXIiOiJwYWJsb25vcmFzQGhvdG1haWwuY29tIn0.k9HllKvDInPs7mMv8WB2cau-c9cMBoo7iPRV97t7f235GEWiLoXDrU_qkFPENnQWEkhyZZdaL5TVSWg7luIShg"
)

var fechaydolararray []Jsonpropio

func main() {

	//var fecha string = ingresarFecha()

	//valoresVariacion := mapadevalores(fydolar)
	//pure(valoresVariacion)

	//Creando una web app
	http.HandleFunc("/", index)

	http.HandleFunc("/dollar", getDollarWeb)
	//Creando una web app

	http.ListenAndServe(":8080", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./index.html")
}

func getDollarWeb(w http.ResponseWriter, r *http.Request) {

	tmpl := template.Must(template.ParseFiles("valorDollar.html"))
	values := r.URL.Query()
	fechaUser := values.Get("dia")

	if strings.Contains(fechaUser, "-") == true {

		arrayfechaUserSinGuion := strings.Split(fechaUser, "-")
		fechaUserSinGuion := arrayfechaUserSinGuion[0] + arrayfechaUserSinGuion[1] + arrayfechaUserSinGuion[2]
		fechaUserInt, _ := strconv.Atoi(fechaUserSinGuion)

		bodyOficial := getOfficialDollarRates("https://api.estadisticasbcra.com/usd_of")
		fydolar := valorestransformados(bodyOficial)
		valordolar := devolvervalordolar(mapadevalores(fydolar), fechaUser)

		mapaDolar2 := mapadevaloresNuevo(fydolar)

		currentTime := time.Now()
		d := currentTime.Format("2006-01-02")
		arraydSinguion := strings.Split(d, "-")
		dSinguion := arraydSinguion[0] + arraydSinguion[1] + arraydSinguion[2]
		dInt, _ := strconv.Atoi(dSinguion)

		//20020302 es el valor de la primera fecha que se tiene valor del dolar, 2002-03-02. Y dInt es el valor entero de la fecha de hoy
		if valordolar == 0 && fechaUserInt >= 20020302 && fechaUserInt <= dInt {
			for valordolar == 0 {
				fechaUserInt = fechaUserInt - 1
				valordolar = mapaDolar2[fechaUserInt]
			}
		}

		bodyOficial = getOfficialDollarRates("https://api.estadisticasbcra.com/usd")
		fydolar = valorestransformados(bodyOficial)
		valordolarblue := devolvervalordolar(mapadevalores(fydolar), fechaUser)
		//idem valordolar2
		mapaDolar2 = mapadevaloresNuevo(fydolar)

		if valordolarblue == 0 && fechaUserInt >= 20020302 && fechaUserInt <= dInt {
			for valordolarblue == 0 {
				fechaUserInt = fechaUserInt - 1
				valordolarblue = mapaDolar2[fechaUserInt]
			}
		}

		if fechaUserInt < 20020302 || fechaUserInt > dInt {
			http.ServeFile(w, r, "./index2.html")
			return
		}

		variacionDolar := getOfficialDollarRates("https://api.estadisticasbcra.com/var_usd_vs_usd_of")
		fydolar = valorestransformados(variacionDolar)
		fechaPure, max := pure(mapadevalores(fydolar))

		result := ValoresFinales{
			DolarOficial: valordolar,
			DolarBlue:    valordolarblue,
			Fecha:        fechaPure,
			Max:          max,
		}

		tmpl.Execute(w, result)
	} else {
		http.ServeFile(w, r, "./index3.html")
		return
	}

}

// Jsonpropio .....
type Jsonpropio struct {
	Fecha string  `json:"d"`
	Valor float64 `json:"v"`
}

//ValoresFinales para ponerlo en el html
type ValoresFinales struct {
	DolarOficial float64
	DolarBlue    float64
	Fecha        string
	Max          float64
}

func getOfficialDollarRates(urlentrada string) []byte {
	url := urlentrada

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Printf("Falló la creación del request a la URL '%s', dando el error %v", url, err.Error())
		os.Exit(1)
	}

	req.Header.Add("Authorization", authToken)

	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Falló el acceso a la URL '%s', dando el error %v", url, err.Error())
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Falló el acceso al body de la respuesta de '%s', dando el error %v", url, err.Error())
		os.Exit(1)
	}

	return body

}

//transformo el chorizo de bytes en mi formato struct Jsonpropio
func valorestransformados(body []byte) []Jsonpropio {
	err := json.Unmarshal(body, &fechaydolararray)
	if err != nil {
		fmt.Println("error:", err)

	}
	return fechaydolararray
}

func mapadevalores(datos []Jsonpropio) map[string]float64 {

	dolar := make(map[string]float64)

	for _, instanciaJSONPropio := range datos {

		dolar[instanciaJSONPropio.Fecha] = instanciaJSONPropio.Valor

	}
	return dolar
}

//Creo un nuevo mapa de valores para usarlos en el if valodolar == 0 que esta en la funcion getDolar
func mapadevaloresNuevo(datos []Jsonpropio) map[int]float64 {

	mapaDolar2 := make(map[int]float64)

	for _, instanciaJSONPropio := range datos {

		arrayfechaInstJSONPropio := strings.Split(instanciaJSONPropio.Fecha, "-")
		fechaInstJSONPropio := arrayfechaInstJSONPropio[0] + arrayfechaInstJSONPropio[1] + arrayfechaInstJSONPropio[2]
		fechaInstJSONPropioInt, _ := strconv.Atoi(fechaInstJSONPropio)

		mapaDolar2[fechaInstJSONPropioInt] = instanciaJSONPropio.Valor

	}
	return mapaDolar2
}

//Para ingresar valor y printearlo en la terminal
/*
func ingresarFecha() string {
	//cargamos los int
	var diaInt int
	var mesInt int
	var añoInt int
	var dia string
	var mes string

	//capturar datos
	fmt.Println("Ingresar Dia:")
	fmt.Scanln(&diaInt)
	fmt.Println("Ingresar Mes:")
	fmt.Scanln(&mesInt)
	fmt.Println("Ingresar año:")
	fmt.Scanln(&añoInt)

	//convertimos a string y si ingresamos valores menores a 10 me pongo el formato de fecha correcto
	if diaInt < 10 {
		dia = "0" + strconv.Itoa(diaInt)
	} else {
		dia = strconv.Itoa(diaInt)
	}
	if mesInt < 10 {
		mes = "0" + strconv.Itoa(mesInt)
	} else {
		mes = strconv.Itoa(mesInt)
	}
	año := strconv.Itoa(añoInt)

	//Estamos cargando los int de dia mes y fech, y lo estamos creando un string

	var Date string = año + "-" + mes + "-" + dia

	return Date

}
*/

func devolvervalordolar(fecha map[string]float64, inputdate string) float64 {
	//var inputdate string = ingresarFecha()
	valordolar := fecha[inputdate]
	return valordolar
}

func devolvervalordolarNuevo(fecha map[int]float64, inputdate int) float64 {
	//var inputdate string = ingresarFecha()
	valordolar2 := fecha[inputdate]
	return valordolar2
}

//pure ...
func pure(maparatios map[string]float64) (string, float64) {

	arrayfechas := make([]string, len(maparatios))

	date := "2020-01-01"

	for f := range maparatios {
		arrayfechas = append(arrayfechas, f)
	}

	sort.Strings(arrayfechas)

	index := sort.SearchStrings(arrayfechas, date)

	max := maparatios[date]
	fechaPure := date

	for i := index; i < len(arrayfechas); i++ {
		if maparatios[arrayfechas[i]] > max {
			max = maparatios[arrayfechas[i]]
			fechaPure = arrayfechas[i]
		}
		return fechaPure, max
	}
	return fechaPure, max
}
