package npmregistry

// PackageInfo represents npm registry package metadata
type PackageInfo struct {
	DistTags DistTags `json:"dist-tags"`
}

// DistTags holds version tags from npm registry
type DistTags struct {
	Latest string `json:"latest"`
}

// Service provides npm registry operations
type Service interface {
	GetLatestVersion(packageName string) (string, error)
}
