package main

import(
		"io/ioutil"
		"http"
		"os"
		"template"
		"fmt"
		"regexp"
		"strings"
		"strconv"
)


type page struct {
	title string
	body []byte
}

type element struct {
	a int
	b int
}

const lenPath = len("/view/")

var templates = make(map[string]*template.Template)

var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")

func getTitle(w http.ResponseWriter, r *http.Request) (title string, err os.Error) {
	title = r.URL.Path[lenPath:]
	if !titleValidator.MatchString(title) {
		http.NotFound(w, r)
		err = os.NewError("Invalid Page Title")
	}
	return
}

func loadPage(title string) (*page, os.Error) {
	filename := "pages/"+ title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &page{title: title, body: body}, nil
}

func init() { 
	for _, tmpl := range []string{"view", "editnash"} { templates[tmpl] = template.MustParseFile(tmpl+".html", nil) 
	} 
}
	
func main() {
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/nash/", nashHandler)
	http.HandleFunc("/calcula/", calculaHandler)
	http.ListenAndServe(":8080", nil)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[lenPath:]
		if !titleValidator.MatchString(title) {
			http.NotFound(w, r)
			return
		}
		fn(w, r, title)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &page{title: title}
	}
	renderTemplate(w, "edit", p)
}


func renderTemplate(w http.ResponseWriter, tmpl string, p *page) {
	err := templates[tmpl].Execute(p, w)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
	}
}

func nashHandler(w http.ResponseWriter, r *http.Request) {
	p, err := loadPage("nash")
	if err != nil {
		p = &page{title: "nash"}
	}
	renderTemplate(w, "editnash", p)
}

func calculaHandler(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	n,k := getMatrixSize(w, body)
	if n*n != k-1  {
		http.Error(w, "Error: La matriu ha de ser quadrada", http.StatusFound)
		return
	}
	
	//Creem la matriu d'equilibris
	matriu := make([][]element, n)
	for i := 0; i < n; i++ {
	  matriu[i] = make([]element, n)
	}
	
	//Convertim l'entrada en matriu d'elements
	cadena := strings.Split(body, "", -1)
	j := 0
	c := 0
	number1 := ""
	number2 := ""

	for i:=0; i<n*n; i++{
		for cadena[j]!="{"{
			j++
		}
		c = j+1
		for cadena[c]!=","{	
			number1+=cadena[c]
			c++
		}
		a,_ := strconv.Atoi(number1)
		number1 = ""
		c++
		for cadena[c]!="}"{
			number2+=cadena[c]
			c++
		}
		b,_ := strconv.Atoi(number2)
		number2 = ""
		j = c+1
		e := element{a: a, b: b}
		matriu[i/n][i%n] = e
	}
	
	
	//Passos 1 2 3
	matriu = nashEquilibrium(w, matriu)

	//Pas 4
	elem, _ := minimax(w, matriu)
	fmt.Fprintf(w, "L'equilibri es %d,%d", elem.a, elem.b)
}

func minimax(w http.ResponseWriter, m [][]element)(element, element){
	minFila := m[0][0].a
	minColumna := m[0][0].b
	
	minCol := make([]int, len(m))
	minFil := make([]int, len(m[0]))
	
	
	
	for i:=0; i<len(m[0]); i++{
		for j:=0; j<len(m); j++{
			if m[i][j].a <= minFila{
				minFil[i] = m[i][j].a
			}
			if m[i][j].b <= minColumna{
				minCol[j] = m[i][j].b
			}
		}
	}
	indexCol:=max(minCol)
	indexFil:=max(minFil)
	
	
	
	return element{a:m[indexFil][indexCol].a, b:m[indexFil][indexCol].b}, element{a:indexFil, b:indexCol}
}

func max(v []int)(int){
	max := v[0]
	index:= -1
	for i:=0; i<len(v);i++{
		if v[i]>=max{
			max = v[i]
			index = i
		}
	}
	return index
	
}

func getMatrixSize(w http.ResponseWriter, input string)(int, int){

	breaks := strings.Split(input, "\n", -1)
	keys := strings.Split(input, "{", -1)
	return len(breaks), len(keys)
	
}

func nashEquilibrium(w http.ResponseWriter, m [][]element)([][]element){
	//Codi de classe de PDGPE
	
	//Estructura per marcar les files o columnes eliminades
	fil := make([]bool, len(m))
	col := make([]bool, len(m))

	numfil := 0
	numcol := 0
	
	pivotfil := 0
	tovipfil := 0
	
	pivotcol := 0
	tovipcol := 0	
	
	for i:=0; i<len(m); i++{
		for j:=0; j<len(m); j++{
			pivotfil=0
			tovipfil=0
			pivotcol=0
			tovipcol=0
			for k:=0; k<len(m); k++{
				//1. Eliminar estrategies dominades per files (M): Fact = {A2:A3}
				if i == j{
					break;
				}
				if m[i][k].a <= m[j][k].a{
					tovipfil++
				} 
				if m[i][k].a > m[j][k].a{
					pivotfil++
				}
				if m[i][k].a == m[j][k].a{
					tovipfil++
					pivotfil++
				}
				if pivotfil == len(m){
					fil[j] = true
				}
				if tovipfil == len(m){
					fil[i] = true
				}
				
				//2. Eliminar estrategies dominades per columnes (M): Cact = {B2:B3}			
				if m[k][i].b <= m[k][j].b{
					tovipcol++
				} 
				if m[k][i].b > m[k][j].b{
					pivotcol++
				}
				if m[k][i].b == m[k][j].b{
					tovipcol++
					pivotcol++
				}
				if pivotcol == len(m){
					col[j] = true
				}
				if tovipcol == len(m){
					col[i] = true
				}				
			}
		}
	}

		
	//Calculem el nombre de files i columnes de la nova matriu
	for i:=0; i<len(m); i++{
		if !fil[i]{
			numfil++
		}
		if !col[i]{
			numcol++
		}
	}
	
	
	//3. M1 = MuntarMatriu(M, Fact, Cact)
	m1 := make([][]element, numcol)
	for i := 0; i < numcol; i++ {
	  m1[i] = make([]element, numfil)
	 }
	
	
	iM := 0;
	jM := 0;
	
	for i:=0; i<numfil; i++{
		for j:=0; j<numcol; j++{	
			if jM == len(m){
				jM = 0
			}
			if fil[iM]{
				iM++
			}
			if col[jM]{
				jM++
			}
			m1[i][j] = m[iM][jM]
			jM++	
		}
		iM++	
	}
	
	
	//4. Si M1 != M llavors M = m1; anar a 1
	//   Si no EO = Estrategia d'equilibri (M1)
	if !iguals(m1, m){
		m1 = nashEquilibrium(w, m1)
	}
	
	return m1
}


func iguals(m1 [][]element, m2 [][]element)(bool){
	return len(m1)==len(m2) && len(m1[0])==len(m2[0])
}




