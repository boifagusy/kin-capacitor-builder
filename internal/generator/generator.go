package generator

import (
	"local-apk-builder/internal/database"
)

// Generate generates all project files
func (g *Generator) Generate() (*GeneratedProject, error) {
	capConfig, err := g.GenerateCapacitorConfig()
	if err != nil {
		return nil, err
	}

	pkgJSON, err := g.GeneratePackageJSON()
	if err != nil {
		return nil, err
	}

	var mainActivity, appBridge string
	switch g.project.RuntimeType {
	case "go-sqlite":
		mainActivity = g.GenerateRuntimeMainActivityJava()
		appBridge = g.GenerateRuntimeAppBridgeJS()
	case "go-native":
		mainActivity = g.GenerateNativeMainActivityJava()
		appBridge = g.GenerateNativeAppBridgeJS()
	default:
		mainActivity = g.GenerateMainActivityJava()
		appBridge = g.GenerateAppBridgeJS()
	}

	return &GeneratedProject{
		ProjectID:          g.project.ID,
		CapacitorConfig:    capConfig,
		PackageJSON:        pkgJSON,
		WebIndex:           g.GenerateWebIndex(),
		AndroidManifest:    g.GenerateAndroidManifest(),
		IOSInfoPlist:       g.GenerateIOSInfoPlist(),
		IconsGenerated:     true,
		SplashGenerated:    true,
		NativeMainActivity: mainActivity,
		AppConfigJSON:      g.GenerateAppConfigJSON(),
		LauncherIconPNG:    g.DecodeLauncherIcon(),
		IconBase64:         g.GetIconBase64(),
		AppBridgeJS:        appBridge,
		GoogleServicesJSON: g.project.GoogleServices,
		SplashPNG:          g.GenerateSplashPNG(),
		ReleaseKeystore:    g.GetReleaseKeystore(),
		StableKeystore:     g.GetStableKeystore(),
	}, nil
}

// GenerateForProject generates for a specific project
func GenerateForProject(project *database.Project) (*GeneratedProject, error) {
	g := NewGenerator(project)
	return g.Generate()
}
