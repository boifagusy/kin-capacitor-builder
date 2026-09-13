package generator

import "local-apk-builder/internal/database"

// GeneratedProject represents the output structure
type GeneratedProject struct {
    ProjectID        int64  `json:"project_id"`
    CapacitorConfig  string `json:"capacitor_config"`
    PackageJSON      string `json:"package_json"`
    WebIndex         string `json:"web_index"`
    AndroidManifest  string `json:"android_manifest"`
    IOSInfoPlist     string `json:"ios_info_plist"`
    IconsGenerated   bool   `json:"icons_generated"`
    SplashGenerated  bool   `json:"splash_generated"`
    NativeMainActivity string `json:"native_main_activity"`
    AppConfigJSON    string `json:"app_config_json"`
    LauncherIconPNG  []byte `json:"launcher_icon_png,omitempty"`
    IconBase64       string `json:"icon_base64,omitempty"`
    AppBridgeJS     string `json:"app_bridge_js"`
    GoogleServicesJSON string `json:"google_services_json"`
    SplashPNG          []byte `json:"splash_png,omitempty"`
    ReleaseKeystore    string `json:"release_keystore,omitempty"`
    StableKeystore     string `json:"stable_keystore,omitempty"`
}

// Generator generates project files from config
type Generator struct {
    project *database.Project
    source  *PreparedSource
}

// NewGenerator creates a generator for a project
func NewGenerator(project *database.Project) *Generator {
    return &Generator{project: project, source: nil}
}

func NewGeneratorWithSource(project *database.Project, source *PreparedSource) *Generator {
    return &Generator{project: project, source: source}
}

// BuildStatus represents build progress
type BuildStatus struct {
    Status    string `json:"status"`
    Step      string `json:"step"`
    Progress  int    `json:"progress"`
    Message   string `json:"message"`
    Error     string `json:"error,omitempty"`
}
