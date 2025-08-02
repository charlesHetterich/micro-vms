package commands

import "fmt"

func (a *App) Delete(ids []string) error {
	records, err := a.Records.Get(ids)
	if err != nil {
		return fmt.Errorf("delete records: %w", err)
	}
	for _, r := range records {
		if err := a.Records.Remove(ids); err != nil {
			return fmt.Errorf("delete record %s: %w", r.ID, err)
		}
	}
	return nil
}
