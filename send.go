package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/emersion/go-vcard"
)

func isVcardFile(filename string) bool {
	return !strings.HasPrefix(filename, ".") && strings.HasSuffix(filename, ".vcf")
}

func vcards(dir string) []string {
	var cards []string

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		filename := file.Name()

		if !isVcardFile(filename) {
			continue
		}

		resolvedPath := path.Join(dir, filename)
		cards = append(cards, resolvedPath)
	}

	return cards
}

func formattedAddress(a *vcard.Address) string {
	if a == nil {
		return ""
	}

	var addr []string

	for _, s := range []string{a.PostOfficeBox, a.ExtendedAddress, a.StreetAddress, a.Locality, a.Region, a.PostalCode, a.Country} {
		if len(s) > 0 {
			addr = append(addr, s)
		}
	}

	return strings.Join(addr, " ")
}

func allCards(path string) []vcard.Card {
	var cards []vcard.Card

	for _, cardPath := range vcards(path) {
		f, err := os.Open(cardPath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		dec := vcard.NewDecoder(f)
		for {
			card, err := dec.Decode()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			cards = append(cards, card)
		}
	}

	return cards
}

func main() {
	var (
		recipientIndex int
		notes          string
	)

	cards := allCards("cards/")
	var recipientOptions []huh.Option[int]
	for i, card := range cards {
		name := card.PreferredValue(vcard.FieldFormattedName)
		recipientOptions = append(recipientOptions, huh.NewOption(name, i))
	}

	// Figure out who to send to
	// Show their address
	// Take in remarks
	// save the output to a file
	// CSV for sent mail
	//	- [Date], [Name], [Address], [Remarks]
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Who are you sending to?").
				Options(recipientOptions...).
				Value(&recipientIndex),
			huh.NewText().Title("Notes").Value(&notes),
		),
	)

	tea.ClearScrollArea()
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	card := cards[recipientIndex]
	a := formattedAddress(card.Address())
	fmt.Println(card.PreferredValue(vcard.FieldFormattedName), "\t", a)
}
