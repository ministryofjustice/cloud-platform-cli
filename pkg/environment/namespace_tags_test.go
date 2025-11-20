package environment

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type mockFileSystem struct {
	files map[string]string
	dirs  map[string]bool
}

func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files: make(map[string]string),
		dirs:  make(map[string]bool),
	}
}

func (m *mockFileSystem) addFile(path, content string) {
	m.files[path] = content
	// Add all parent directories
	dir := filepath.Dir(path)
	for dir != "." && dir != "/" {
		m.dirs[dir] = true
		dir = filepath.Dir(dir)
	}
}

func Test_newTagChecker(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		want    *TagChecker
	}{
		{
			name:    "Creates TagChecker with correct baseDir and searchTags",
			baseDir: "/path/to/repo",
			want: &TagChecker{
				searchTags: []string{"business-unit", "application", "is-production", "owner", "namespace"},
				baseDir:    "/path/to/repo",
			},
		},
		{
			name:    "Creates TagChecker with different path",
			baseDir: "/another/path",
			want: &TagChecker{
				searchTags: []string{"business-unit", "application", "is-production", "owner", "namespace"},
				baseDir:    "/another/path",
			},
		},
		{
			name:    "Creates TagChecker with empty path",
			baseDir: "",
			want: &TagChecker{
				searchTags: []string{"business-unit", "application", "is-production", "owner", "namespace"},
				baseDir:    "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newTagChecker(tt.baseDir)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTagChecker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagChecker_checkAndAddTags(t *testing.T) {
	tests := []struct {
		name      string
		baseDir   string
		namespace string
		wantErr   bool
	}{
		{
			name:      "Non-existent namespace",
			baseDir:   "/path/to/repo",
			namespace: "nonexistent-namespace",
			wantErr:   false,
		},
		{
			name:      "Existing namespace with all tags",
			baseDir:   "/path/to/repo",
			namespace: "existing-namespace-with-tags",
			wantErr:   false,
		},
		{
			name:      "Existing namespace missing some tags",
			baseDir:   "/path/to/repo",
			namespace: "existing-namespace-missing-tags",
			wantErr:   false,
		},
		{
			name:      "Invalid base directory",
			baseDir:   "/invalid/path",
			namespace: "some-namespace",
			wantErr:   false,
		},
		{
			name:      "Empty namespace",
			baseDir:   "/path/to/repo",
			namespace: "",
			wantErr:   false,
		},
		{
			name:      "Namespace with special characters",
			baseDir:   "/path/to/repo",
			namespace: "namespace-!@#$%^&*()",
			wantErr:   false,
		},
		{
			name:      "Very long namespace",
			baseDir:   "/path/to/repo",
			namespace: "this-is-a-very-long-namespace-name-to-test-the-functionality-of-the-tag-checker-in-handling-long-strings",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTagChecker(tt.baseDir)
			gotErr := tc.checkAndAddTags(tt.namespace)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("checkAndAddTags() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("checkAndAddTags() succeeded unexpectedly")
			}
		})
	}
}

func TestTagChecker_findTerraformFile(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		mockSetup  func() *mockFileSystem
		wantPath   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "Valid namespace with existing terraform file",
			namespace: "existing-namespace",
			mockSetup: func() *mockFileSystem {
				mfs := newMockFileSystem()
				path := "/mock/repo/namespaces/live.cloud-platform.service.justice.gov.uk/existing-namespace/resources/main.tf"
				mfs.addFile(path, "# terraform config")
				return mfs
			},
			wantPath: "/mock/repo/namespaces/live.cloud-platform.service.justice.gov.uk/existing-namespace/resources/main.tf",
			wantErr:  false,
		},
		{
			name:      "Namespace in live-2 cluster",
			namespace: "test-namespace",
			mockSetup: func() *mockFileSystem {
				mfs := newMockFileSystem()
				path := "/mock/repo/namespaces/live-2.cloud-platform.service.justice.gov.uk/test-namespace/resources/main.tf"
				mfs.addFile(path, "# terraform config")
				return mfs
			},
			wantPath: "/mock/repo/namespaces/live-2.cloud-platform.service.justice.gov.uk/test-namespace/resources/main.tf",
			wantErr:  false,
		},
		{
			name:      "Non-existent namespace",
			namespace: "nonexistent-namespace",
			mockSetup: func() *mockFileSystem {
				return newMockFileSystem()
			},
			wantErr:    true,
			wantErrMsg: "no terraform file found for namespace",
		},
		{
			name:      "Empty namespace",
			namespace: "",
			mockSetup: func() *mockFileSystem {
				return newMockFileSystem()
			},
			wantErr:    true,
			wantErrMsg: "no terraform file found for namespace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for actual file system test
			tmpDir := t.TempDir()
			mfs := tt.mockSetup()

			// Create actual files based on mock
			for path := range mfs.files {
				// Replace /mock/repo with tmpDir
				actualPath := strings.Replace(path, "/mock/repo", tmpDir, 1)
				dir := filepath.Dir(actualPath)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(actualPath, []byte(mfs.files[path]), 0o644); err != nil {
					t.Fatalf("Failed to write file: %v", err)
				}
			}

			tc := newTagChecker(tmpDir)
			got, gotErr := tc.findTerraformFile(tt.namespace)

			if tt.wantErr {
				if gotErr == nil {
					t.Fatal("findTerraformFile() succeeded unexpectedly")
				}
				if tt.wantErrMsg != "" && !strings.Contains(gotErr.Error(), tt.wantErrMsg) {
					t.Errorf("findTerraformFile() error = %v, want error containing %q", gotErr, tt.wantErrMsg)
				}
				return
			}

			if gotErr != nil {
				t.Errorf("findTerraformFile() failed: %v", gotErr)
				return
			}

			// Adjust expected path
			expectedPath := strings.Replace(tt.wantPath, "/mock/repo", tmpDir, 1)
			if got != expectedPath {
				t.Errorf("findTerraformFile() = %v, want %v", got, expectedPath)
			}
		})
	}
}

func TestTagChecker_findTerraformFileRecursive(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		mockSetup  func() *mockFileSystem
		wantPath   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:      "Valid namespace with standard structure",
			namespace: "existing-namespace",
			mockSetup: func() *mockFileSystem {
				mfs := newMockFileSystem()
				path := "/mock/repo/namespaces/live.cloud-platform.service.justice.gov.uk/existing-namespace/resources/main.tf"
				mfs.addFile(path, "# terraform config")
				return mfs
			},
			wantPath: "/mock/repo/namespaces/live.cloud-platform.service.justice.gov.uk/existing-namespace/resources/main.tf",
			wantErr:  false,
		},
		{
			name:      "Valid namespace with custom nested path",
			namespace: "custom-namespace",
			mockSetup: func() *mockFileSystem {
				mfs := newMockFileSystem()
				path := "/mock/repo/namespaces/custom/deep/path/custom-namespace/resources/main.tf"
				mfs.addFile(path, "# custom terraform")
				return mfs
			},
			wantPath: "/mock/repo/namespaces/custom/deep/path/custom-namespace/resources/main.tf",
			wantErr:  false,
		},
		{
			name:      "Multiple namespaces, finds correct one",
			namespace: "target-namespace",
			mockSetup: func() *mockFileSystem {
				mfs := newMockFileSystem()
				mfs.addFile("/mock/repo/namespaces/live/other-namespace/resources/main.tf", "# other")
				mfs.addFile("/mock/repo/namespaces/live/target-namespace/resources/main.tf", "# target")
				return mfs
			},
			wantPath: "/mock/repo/namespaces/live/target-namespace/resources/main.tf",
			wantErr:  false,
		},
		{
			name:      "Non-existent namespace",
			namespace: "nonexistent-namespace",
			mockSetup: func() *mockFileSystem {
				mfs := newMockFileSystem()
				// Add some other namespaces but not the target
				mfs.addFile("/mock/repo/namespaces/live/other-namespace/resources/main.tf", "# other")
				return mfs
			},
			wantErr:    true,
			wantErrMsg: "no terraform file found for namespace",
		},
		{
			name:      "Empty namespace",
			namespace: "",
			mockSetup: func() *mockFileSystem {
				return newMockFileSystem()
			},
			wantErr:    true,
			wantErrMsg: "no terraform file found for namespace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for actual file system test
			tmpDir := t.TempDir()
			mfs := tt.mockSetup()

			// Create actual files based on mock
			for path := range mfs.files {
				// Replace /mock/repo with tmpDir
				actualPath := strings.Replace(path, "/mock/repo", tmpDir, 1)
				dir := filepath.Dir(actualPath)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
				if err := os.WriteFile(actualPath, []byte(mfs.files[path]), 0o644); err != nil {
					t.Fatalf("Failed to write file: %v", err)
				}
			}

			tc := newTagChecker(tmpDir)
			got, gotErr := tc.findTerraformFileRecursive(tt.namespace)

			if tt.wantErr {
				if gotErr == nil {
					t.Fatal("findTerraformFileRecursive() succeeded unexpectedly")
				}
				if tt.wantErrMsg != "" && !strings.Contains(gotErr.Error(), tt.wantErrMsg) {
					t.Errorf("findTerraformFileRecursive() error = %v, want error containing %q", gotErr, tt.wantErrMsg)
				}
				return
			}

			if gotErr != nil {
				t.Errorf("findTerraformFileRecursive() failed: %v", gotErr)
				return
			}

			// Adjust expected path
			expectedPath := strings.Replace(tt.wantPath, "/mock/repo", tmpDir, 1)
			if got != expectedPath {
				t.Errorf("findTerraformFileRecursive() = %v, want %v", got, expectedPath)
			}
		})
	}
}

func TestTagChecker_findMissingTags(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name: "All tags present in provider default_tags",
			content: `provider "aws" {
		region = "eu-west-2"
		default_tags {
			tags = {
				business-unit  = var.business_unit
				application    = var.application
				is-production  = var.is_production
				owner          = var.team_name
				namespace      = var.namespace
			}
		}
	}`,
			want: []string{},
		},
		{
			name: "All tags present in locals block",
			content: `locals {
		default_tags = {
			business-unit  = var.business_unit
			application    = var.application
			is-production  = var.is_production
			owner          = var.team_name
			namespace      = var.namespace
		}
	}`,
			want: []string{},
		},
		{
			name: "Some tags missing in default_tags",
			content: `provider "aws" {
		region = "eu-west-2"
		default_tags {
			tags = {
				business-unit  = var.business_unit
				owner          = var.team_name
			}
		}
	}`,
			want: []string{"application", "is-production", "namespace"},
		},
		{
			name: "No default_tags block present",
			content: `provider "aws" {
		region = "eu-west-2"
	}`,
			want: []string{"business-unit", "application", "is-production", "owner", "namespace"},
		},
		{
			name:    "Empty content",
			content: ``,
			want:    []string{"business-unit", "application", "is-production", "owner", "namespace"},
		},
		{
			name: "Tags with quoted keys",
			content: `provider "aws" {
		default_tags {
			tags = {
				"business-unit"  = var.business_unit
				"application"    = var.application
				"is-production"  = var.is_production
				"owner"          = var.team_name
				"namespace"      = var.namespace
			}
		}
	}`,
			want: []string{},
		},
		{
			name: "Mixed - some quoted, some missing",
			content: `provider "aws" {
		default_tags {
			tags = {
				"business-unit"  = var.business_unit
				application      = var.application
			}
		}
	}`,
			want: []string{"is-production", "owner", "namespace"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTagChecker("/test")
			got := tc.findMissingTags(tt.content)

			// Compare lengths first
			if len(got) != len(tt.want) {
				t.Errorf("findMissingTags() returned %d tags, want %d\nGot:  %v\nWant: %v", len(got), len(tt.want), got, tt.want)
				return
			}

			// Check each expected tag is in the result
			for _, wantTag := range tt.want {
				found := false
				for _, gotTag := range got {
					if gotTag == wantTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("findMissingTags() missing expected tag %q\nGot:  %v\nWant: %v", wantTag, got, tt.want)
				}
			}
		})
	}
}

func TestTagChecker_addMissingTags(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		missingTags []string
		wantErr     bool
	}{
		{
			name: "Add missing tags to existing default_tags block",
			content: `provider "aws" {
  region = "eu-west-2"
  default_tags {
    tags = {
      business-unit  = var.business_unit
      owner          = var.team_name
    }
  }
}`,
			missingTags: []string{"application", "is-production", "namespace"},
			wantErr:     false,
		},
		{
			name: "Create default_tags block when missing",
			content: `provider "aws" {
  region = "eu-west-2"
}`,
			missingTags: []string{"business-unit", "application", "is-production", "owner", "namespace"},
			wantErr:     false,
		},
		{
			name: "No missing tags - should not modify file",
			content: `provider "aws" {
  default_tags {
    tags = {
      business-unit  = var.business_unit
      application    = var.application
      is-production  = var.is_production
      owner          = var.team_name
      namespace      = var.namespace
    }
  }
}`,
			missingTags: []string{},
			wantErr:     false,
		},
		{
			name: "Add single missing tag",
			content: `provider "aws" {
  default_tags {
    tags = {
      business-unit  = var.business_unit
      application    = var.application
      is-production  = var.is_production
      owner          = var.team_name
    }
  }
}`,
			missingTags: []string{"namespace"},
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and file
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "main.tf")

			// Create the file with initial content
			dir := filepath.Dir(filePath)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("Failed to create directory: %v", err)
			}
			if err := os.WriteFile(filePath, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("Failed to write initial file: %v", err)
			}

			tc := newTagChecker(tmpDir)
			gotErr := tc.addMissingTags(filePath, tt.content, tt.missingTags)

			if tt.wantErr {
				if gotErr == nil {
					t.Fatal("addMissingTags() succeeded unexpectedly")
				}
				return
			}

			if gotErr != nil {
				t.Errorf("addMissingTags() failed: %v", gotErr)
				return
			}

			// Verify file was written
			modifiedContent, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("Failed to read modified file: %v", err)
			}

			// Verify all missing tags were added
			contentStr := string(modifiedContent)
			for _, tag := range tt.missingTags {
				if !strings.Contains(contentStr, tag) {
					t.Errorf("addMissingTags() modified content missing tag %q\nContent:\n%s", tag, contentStr)
				}
			}
		})
	}
}

func TestTagChecker_addDefaultTagsBlock(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		lines   []string
		want    []string
	}{
		{
			name:    "Adds default_tags block to provider",
			baseDir: "/test",
			lines: []string{
				`provider "aws" {`,
				`  region = "eu-west-2"`,
				`}`,
			},
			want: []string{
				`provider "aws" {`,
				`  default_tags {`,
				`    tags = {`,
				`      business-unit = var.business_unit`,
				`      application = var.application`,
				`      is-production = var.is_production`,
				`      owner = var.team_name`,
				`      namespace = var.namespace`,
				`    }`,
				`  }`,
				`  region = "eu-west-2"`,
				`}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTagChecker(tt.baseDir)
			got := tc.addDefaultTagsBlock(tt.lines)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addDefaultTagsBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTagChecker_getTagValue(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		tag     string
		want    string
	}{
		{
			name:    "Get value for business-unit tag",
			baseDir: "/test",
			tag:     "business-unit",
			want:    "var.business_unit",
		},
		{
			name:    "Get value for is-production tag",
			baseDir: "/test",
			tag:     "is-production",
			want:    "var.is_production",
		},
		{
			name:    "Get value for owner tag",
			baseDir: "/test",
			tag:     "owner",
			want:    "var.team_name",
		},
		{
			name:    "Get value for application tag",
			baseDir: "/test",
			tag:     "application",
			want:    "var.application",
		},
		{
			name:    "Get value for namespace tag",
			baseDir: "/test",
			tag:     "namespace",
			want:    "var.namespace",
		},
		{
			name:    "Get value for custom tag",
			baseDir: "/test",
			tag:     "custom-tag",
			want:    "var.custom_tag",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTagChecker(tt.baseDir)
			got := tc.getTagValue(tt.tag)
			if got != tt.want {
				t.Errorf("getTagValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
