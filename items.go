package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
)

type item struct {
	Title   string
	Items   []item
	Content string
}

func initList() list.Model {

	sessionList := getActiveSessions()

	if len(sessionList) > 0 {
		// Add an item "Running Sessions" in the menu
		menu.Items = append(menu.Items, item{
			Title:   "Running Sessions",
			Items:   nil,                // Sub-items will be added dynamically
			Content: "Commands running", // Add content if needed
		})

		// For each session, add a subitem to the "Running Sessions" item
		for _, session := range sessionList {
			menu.Items[len(menu.Items)-1].Items = append(menu.Items[len(menu.Items)-1].Items, item{
				Title:   session.tool + " - " + session.description,
				Items:   nil, // Sub-sub-items if needed
				Content: session.uuid,
			})
		}
	}

	items := menu

	const defaultWidth = 20

	l := list.New(convertToItems(items.Items), itemDelegate{}, defaultWidth, listHeight)
	l.Title = items.Title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	return l
}

func initItemsListFile() item {
	items, err := getItemsFromFile("scan.conf.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}
	// Check if there is a '-' in the title
	for _, subItem := range items.Items {

		if strings.Contains(subItem.Title, "-") {
			// Handle the case where '-' is present in the title
			fmt.Printf("Warning: Title '%s' contains '-' character.\n", subItem.Title)
			// quit
			os.Exit(1)
		}
		for _, subSubItem := range subItem.Items {
			if strings.Contains(subSubItem.Title, "-") {
				// Handle the case where '-' is present in the title
				fmt.Printf("Warning: Title '%s' contains '-' character.\n", subSubItem.Title)
				// quit
				os.Exit(1)
			}

		}
	}

	return items
}

func getItemsFromFile(filename string) (item, error) {
	file, err := os.Open(filename)
	if err != nil {
		return item{}, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return item{}, err
	}

	var items item
	err = json.Unmarshal(data, &items)
	if err != nil {
		return item{}, err
	}

	return items, nil
}

func convertToItems(items []item) []list.Item {
	var result []list.Item
	for _, i := range items {
		result = append(result, i)
	}
	return result
}
