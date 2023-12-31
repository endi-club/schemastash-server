package types

type Schematic struct {
	ID            string
	LatestVersion string
	CreatedAt     string
	Versions      map[string]Version
	Data          string // Master data

	Archived   bool
	ArchivedAt string

	Description string
}

type Version struct {
	ID          string
	SchematicID string
	Data        string
	CreatedAt   string
}
