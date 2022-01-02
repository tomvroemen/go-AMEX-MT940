package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"os/user"
	"sort"
	"strings"
)

func main() {

	entries := map[string][5]string{}
	keys := []string{}

	fmt.Println("Convert your AMEX csv export files into an MT940 file.")
	// AMEX downloads to a file named "ofx.csv"
	// we are on MAC OSX
	usr, err := user.Current()
	if err != nil {
		fmt.Println("could'nt establish user home directory", err)
	}
	fmt.Println(usr.HomeDir)

	csvfile, err := os.Open(usr.HomeDir + "/Downloads/ofx.csv")
	if err != nil {
		csvfile, err = os.Open(usr.HomeDir + "/Downloads/activity.csv")
		if err != nil {
			fmt.Println("Problems opening the file", err)
			return
		}
	}

	// Parse
	r := csv.NewReader(csvfile)

	// Loop over the lines
	for {
		// Read line record from csv
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
		}
		if len(line) < 5 {
			fmt.Println("incomplete line")
			continue
		}
		line[0] = strings.Replace(line[0], "/", "-", -1)
		entryKey := strings.Split(line[0], "-")[2] + "-" + strings.Split(line[0], "-")[1] + "-" + strings.Split(line[0], "-")[0] + "-" + line[1]

		keys = append(keys, entryKey)

		entries[entryKey] = [5]string{line[0], line[1], line[2], line[3], line[4]}
	}

	sort.Strings(keys)

	// Make sure the user has the correct file
	fmt.Println("First entry:", entries[keys[0]][0])
	fmt.Println("Last  entry:", entries[keys[len(keys)-1]][0])

	// Take input from user
	var iban string
	fmt.Println("what is your IBAN number to write these transactions to?")
	fmt.Scanln(&iban)

	if strings.ToLower(iban) != "" {
		fmt.Println("exporting MT-940")

		result := ""

		for _, v := range entries { // 31-12-20,"Reference: AT210010032000010023035"," 99,00","APPLE.COM/NL HOLLYHILL"," Datum transactie verwerkt 31-12-20",
			dateCode := strings.Split(v[0], "-")

			result += `
{1:F01INGBNL2ABXXX0000000000}
{2:I940INGBNL2AXXXN}
{4:
:20:P210404000000001
:25:` + iban + `EUR
:28C:00000
:60F:D` + dateCode[2] + dateCode[1] + dateCode[0] + `EUR0,00
:61:` + dateCode[2] + dateCode[1] + dateCode[0] + `D` + v[2] + `NTRFNONREF//` + dateCode[2] + dateCode[1] + dateCode[0] + `00000001
/TRCD/00100/
:86:/CNTP///` + v[3] + `///REMI/USTD//AMEX ` + v[1] + ` ` + v[3] + `/
:62F:D` + dateCode[2] + dateCode[1] + dateCode[0] + `EUR` + v[4] + `
:64:D` + dateCode[2] + dateCode[1] + dateCode[0] + `EUR` + v[4] + `
:65:D` + dateCode[2] + dateCode[1] + dateCode[0] + `EUR` + v[4] + `
:65:D` + dateCode[2] + dateCode[1] + dateCode[0] + `EUR` + v[4] + `
:86:/SUM/1/0/` + v[4] + `/0,00/
-}
			`
		}
		//fmt.Println(result)
		f, err := os.Create(usr.HomeDir + "/Downloads/" + iban + ".txt")

		if err != nil {
			fmt.Println(err)
		}

		defer f.Close()

		_, err2 := f.WriteString(result)

		if err2 != nil {
			fmt.Println(err2)
		}
	} else {
		fmt.Println("OK. stopping.")
	}

	fmt.Println("Done!")
	return
}
