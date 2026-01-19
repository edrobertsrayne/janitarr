package forms

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

// ConfirmDelete shows a confirmation dialog for destructive delete operations.
// Requires the user to type the item name to confirm.
func ConfirmDelete(itemType, itemName string) (bool, error) {
	var confirmation string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprintf("Delete %s", itemType)).
				Description(fmt.Sprintf("Are you sure you want to delete '%s'?", itemName)),

			huh.NewInput().
				Title("Type the name to confirm").
				Description(fmt.Sprintf("Type '%s' to confirm deletion", itemName)).
				Validate(func(s string) error {
					if strings.TrimSpace(s) != itemName {
						return fmt.Errorf("name does not match")
					}
					return nil
				}).
				Value(&confirmation),
		),
	).WithTheme(huh.ThemeBase())

	err := form.Run()
	if err != nil {
		// User cancelled (Escape key)
		return false, nil
	}

	return confirmation == itemName, nil
}

// ConfirmAction shows a simple yes/no confirmation dialog
func ConfirmAction(message string) (bool, error) {
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(message).
				Description("This action cannot be undone").
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeBase())

	err := form.Run()
	if err != nil {
		// User cancelled
		return false, nil
	}

	return confirmed, nil
}

// ConfirmActionWithDetails shows a confirmation with additional details
func ConfirmActionWithDetails(title, details string) (bool, error) {
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(title).
				Description(details),

			huh.NewConfirm().
				Title("Proceed?").
				Description("This action cannot be undone").
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeBase())

	err := form.Run()
	if err != nil {
		// User cancelled
		return false, nil
	}

	return confirmed, nil
}
