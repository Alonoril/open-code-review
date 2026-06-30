package reviewbundle

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
)

// PreparePartitioned builds a deterministic diff manifest whose bundle parts
// each obey PrepareOptions.MaxBundleSize.
func PreparePartitioned(
	ctx context.Context,
	options PrepareOptions,
) (*ScanManifest, []byte, error) {
	maxBundleSize := options.MaxBundleSize
	if maxBundleSize <= 0 {
		maxBundleSize = DefaultMaxBundleBytes
	}
	fullOptions := options
	fullOptions.MaxBundleSize = math.MaxInt64
	full, _, err := Prepare(ctx, fullOptions)
	if err != nil {
		return nil, nil, err
	}
	manifest := &ScanManifest{
		SchemaVersion: ScanManifestSchemaVersion,
		Root:          options.RepoDir,
		TargetHash:    full.Target.DiffSHA256,
		BatchStrategy: "diff",
		BatchSize:     1,
		Summary:       full.Summary,
		SkippedFiles:  make([]ScanSkippedFile, 0),
		Bundles:       make([]Bundle, 0),
	}
	var current []File
	for _, file := range full.Files {
		candidate := append(append([]File(nil), current...), file)
		_, encoded, buildErr := buildDiffPartition(full, candidate, maxBundleSize)
		if buildErr != nil {
			return nil, nil, buildErr
		}
		if int64(len(encoded)) <= maxBundleSize {
			current = candidate
			continue
		}
		if len(current) == 0 {
			return nil, nil, singleFilePartitionError(file.Path, len(encoded), maxBundleSize)
		}
		previous, _, buildErr := buildDiffPartition(full, current, maxBundleSize)
		if buildErr != nil {
			return nil, nil, buildErr
		}
		manifest.Bundles = append(manifest.Bundles, *previous)
		current = []File{file}
		_, singleEncoded, buildErr := buildDiffPartition(full, current, maxBundleSize)
		if buildErr != nil {
			return nil, nil, buildErr
		}
		if int64(len(singleEncoded)) > maxBundleSize {
			return nil, nil, singleFilePartitionError(file.Path, len(singleEncoded), maxBundleSize)
		}
	}
	if len(current) > 0 {
		bundle, _, buildErr := buildDiffPartition(full, current, maxBundleSize)
		if buildErr != nil {
			return nil, nil, buildErr
		}
		manifest.Bundles = append(manifest.Bundles, *bundle)
	}
	manifestID, err := computeManifestID(manifest)
	if err != nil {
		return nil, nil, err
	}
	manifest.ManifestID = manifestID
	encoded, err := marshalManifest(manifest)
	if err != nil {
		return nil, nil, err
	}
	return manifest, encoded, nil
}

func buildDiffPartition(
	full *Bundle,
	files []File,
	maxBundleSize int64,
) (*Bundle, []byte, error) {
	partition := &Bundle{
		SchemaVersion:  BundleSchemaVersion,
		Target:         full.Target,
		WorkspaceState: full.WorkspaceState,
		Rules:          make(map[string]Rule),
		Files:          append([]File(nil), files...),
		Contract:       DefaultContract(),
		Warnings: []ProtocolNotice{{
			Code: "diff_partition", Message: "deterministic large-diff partition",
		}},
	}
	partition.Contract.MaxBundleBytes = maxBundleSize
	for _, file := range files {
		partition.Summary.TotalFiles++
		partition.Summary.Insertions += file.Insertions
		partition.Summary.Deletions += file.Deletions
		if file.Reviewable {
			partition.Summary.ReviewableFiles++
		} else {
			partition.Summary.ExcludedFiles++
		}
		if rule, exists := full.Rules[file.RuleID]; exists {
			partition.Rules[file.RuleID] = rule
		}
	}
	bundleID, err := computeBundleID(partition)
	if err != nil {
		return nil, nil, err
	}
	partition.BundleID = bundleID
	encoded, err := marshalWithStableSize(partition)
	if err != nil {
		return nil, nil, err
	}
	return partition, encoded, nil
}

func singleFilePartitionError(path string, size int, maximum int64) error {
	return &ProtocolError{
		Code: "bundle_too_large",
		Message: fmt.Sprintf(
			"file %s requires a %d-byte bundle; maximum is %d",
			path,
			size,
			maximum,
		),
	}
}

func marshalManifest(manifest *ScanManifest) ([]byte, error) {
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return nil, fmt.Errorf("marshal review manifest: %w", err)
	}
	return encoded, nil
}
