package domain

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"unicode/utf8"
)

var allowedTargetArchitectureStyles = map[string]struct{}{
	"modular_monolith":   {},
	"clean_architecture": {},
	"hexagonal":          {},
}

var allowedTargetCompatibilityModes = map[string]struct{}{
	"strict":           {},
	"approved_changes": {},
}

var envSecretRefRegex = regexp.MustCompile(`^env:[A-Z_][A-Z0-9_]*$`)

// TargetEnvironmentError describes an invalid target environment specification record without echoing private contents.
type TargetEnvironmentError struct {
	Reason string
	Cause  error
}

// Error identifies the failed target environment obligation.
func (e *TargetEnvironmentError) Error() string {
	return "invalid target environment: " + e.Reason
}

// Unwrap preserves JSON and decoding failures for structured inspection.
func (e *TargetEnvironmentError) Unwrap() error { return e.Cause }

// ValidateTargetIntent is a semantic alias for ValidateTargetEnvironment.
var ValidateTargetIntent = ValidateTargetEnvironment

// ValidateTargetEnvironment validates a user-owned target environment specification record
// conforming to specs/project/target.schema.json (version 1 or version 2.0).
// In version 2.0, runtime (version, os, architecture) and deployment (model, profile_ref)
// are required, and framework="none" requires version=null while framework!="none" requires
// version to be a non-empty string.
func ValidateTargetEnvironment(raw []byte) error {
	record, err := parseTargetEnvironmentJSON(raw)
	if err != nil {
		return err
	}

	isV2 := false
	if verRaw, present := record["schema_version"]; present {
		var ver string
		if string(verRaw) == "null" || json.Unmarshal(verRaw, &ver) != nil || ver != "2.0" {
			return &TargetEnvironmentError{Reason: "unsupported schema_version"}
		}
		isV2 = true
	}

	requiredFields := []string{
		"language",
		"framework",
		"version",
		"architecture",
		"database_policy",
		"path",
	}
	if isV2 {
		requiredFields = append(requiredFields, "schema_version", "runtime", "deployment")
	}

	for _, f := range requiredFields {
		if _, present := record[f]; !present {
			return &TargetEnvironmentError{Reason: "missing required field: " + f}
		}
	}

	allowedOptionalFields := map[string]struct{}{
		"runtime":         {},
		"web":             {},
		"database":        {},
		"cache":           {},
		"queue":           {},
		"storage":         {},
		"package_manager": {},
		"compatibility":   {},
	}
	if isV2 {
		allowedOptionalFields["deployment"] = struct{}{}
		allowedOptionalFields["schema_version"] = struct{}{}
	}

	reqMap := make(map[string]struct{}, len(requiredFields))
	for _, f := range requiredFields {
		reqMap[f] = struct{}{}
	}

	for k := range record {
		if _, req := reqMap[k]; req {
			continue
		}
		if _, opt := allowedOptionalFields[k]; opt {
			continue
		}
		return &TargetEnvironmentError{Reason: "unexpected field: " + k}
	}

	var language string
	if string(record["language"]) == "null" || json.Unmarshal(record["language"], &language) != nil || len(language) < 1 {
		return &TargetEnvironmentError{Reason: "invalid language"}
	}

	var framework string
	if string(record["framework"]) == "null" || json.Unmarshal(record["framework"], &framework) != nil || len(framework) < 1 {
		return &TargetEnvironmentError{Reason: "invalid framework"}
	}

	if isV2 {
		if framework == "none" {
			if string(record["version"]) != "null" {
				return &TargetEnvironmentError{Reason: "version must be null when framework is none"}
			}
		} else {
			var ver string
			if string(record["version"]) == "null" || json.Unmarshal(record["version"], &ver) != nil || len(ver) < 1 {
				return &TargetEnvironmentError{Reason: "version must be non-empty string when framework is not none"}
			}
		}
	} else {
		var ver string
		if string(record["version"]) == "null" || json.Unmarshal(record["version"], &ver) != nil || len(ver) < 1 {
			return &TargetEnvironmentError{Reason: "invalid version"}
		}
	}

	var dbPolicy string
	if string(record["database_policy"]) == "null" || json.Unmarshal(record["database_policy"], &dbPolicy) != nil || dbPolicy != "preserve_existing" {
		return &TargetEnvironmentError{Reason: "database_policy must be preserve_existing"}
	}

	var path string
	if string(record["path"]) == "null" || json.Unmarshal(record["path"], &path) != nil || len(path) < 1 {
		return &TargetEnvironmentError{Reason: "invalid path"}
	}

	if err := validateTargetArchitecture(record["architecture"]); err != nil {
		return err
	}

	if isV2 || record["runtime"] != nil {
		if err := validateTargetRuntime(record["runtime"], isV2); err != nil {
			return err
		}
	}

	if isV2 {
		if err := validateTargetDeployment(record["deployment"]); err != nil {
			return err
		}
	}

	if record["web"] != nil {
		if err := validateTargetWeb(record["web"]); err != nil {
			return err
		}
	}

	if record["database"] != nil {
		if err := validateTargetDatabase(record["database"]); err != nil {
			return err
		}
	}

	if record["cache"] != nil {
		if err := validateTargetCacheOrQueue(record["cache"], "cache"); err != nil {
			return err
		}
	}

	if record["queue"] != nil {
		if err := validateTargetCacheOrQueue(record["queue"], "queue"); err != nil {
			return err
		}
	}

	if record["storage"] != nil {
		if err := validateTargetStorage(record["storage"]); err != nil {
			return err
		}
	}

	if record["package_manager"] != nil {
		if err := validateTargetPackageManager(record["package_manager"]); err != nil {
			return err
		}
	}

	if record["compatibility"] != nil {
		if err := validateTargetCompatibility(record["compatibility"]); err != nil {
			return err
		}
	}

	return nil
}

func validateTargetArchitecture(raw json.RawMessage) error {
	var arch map[string]json.RawMessage
	if string(raw) == "null" || json.Unmarshal(raw, &arch) != nil || arch == nil {
		return &TargetEnvironmentError{Reason: "architecture must be an object"}
	}
	var style string
	if string(arch["style"]) == "null" || json.Unmarshal(arch["style"], &style) != nil {
		return &TargetEnvironmentError{Reason: "architecture style is required"}
	}
	if _, ok := allowedTargetArchitectureStyles[style]; !ok {
		return &TargetEnvironmentError{Reason: "invalid architecture style enum"}
	}
	if sl, present := arch["service_layer"]; present && string(sl) != "null" {
		var serviceLayer string
		if json.Unmarshal(sl, &serviceLayer) != nil || len(serviceLayer) < 1 {
			return &TargetEnvironmentError{Reason: "invalid architecture service_layer"}
		}
	}
	for k := range arch {
		if k != "style" && k != "service_layer" {
			return &TargetEnvironmentError{Reason: "unexpected field in architecture: " + k}
		}
	}
	return nil
}

func validateTargetRuntime(raw json.RawMessage, isV2 bool) error {
	if string(raw) == "null" {
		if isV2 {
			return &TargetEnvironmentError{Reason: "runtime is required in version 2.0"}
		}
		return nil
	}
	var rt map[string]json.RawMessage
	if json.Unmarshal(raw, &rt) != nil || rt == nil {
		return &TargetEnvironmentError{Reason: "runtime must be an object"}
	}

	allowedRuntimeFields := map[string]struct{}{
		"version":      {},
		"image_digest": {},
		"secret_refs":  {},
		"os":           {},
		"architecture": {},
	}
	for k := range rt {
		if _, ok := allowedRuntimeFields[k]; !ok {
			return &TargetEnvironmentError{Reason: "unexpected field in runtime: " + k}
		}
	}

	if isV2 {
		for _, req := range []string{"version", "os", "architecture"} {
			if _, present := rt[req]; !present || string(rt[req]) == "null" {
				return &TargetEnvironmentError{Reason: "missing required runtime field: " + req}
			}
		}
	}

	for _, strField := range []string{"version", "image_digest", "os", "architecture"} {
		if val, present := rt[strField]; present && string(val) != "null" {
			var s string
			if json.Unmarshal(val, &s) != nil || len(s) < 1 {
				return &TargetEnvironmentError{Reason: "invalid runtime " + strField}
			}
		}
	}

	if sr, present := rt["secret_refs"]; present && string(sr) != "null" {
		var secrets map[string]string
		if json.Unmarshal(sr, &secrets) != nil || secrets == nil {
			return &TargetEnvironmentError{Reason: "invalid runtime secret_refs"}
		}
		for _, ref := range secrets {
			if !envSecretRefRegex.MatchString(ref) {
				return &TargetEnvironmentError{Reason: "invalid secret_refs value format"}
			}
		}
	}
	return nil
}

func validateTargetDeployment(raw json.RawMessage) error {
	if string(raw) == "null" {
		return &TargetEnvironmentError{Reason: "deployment is required in version 2.0"}
	}
	var dep map[string]json.RawMessage
	if json.Unmarshal(raw, &dep) != nil || dep == nil {
		return &TargetEnvironmentError{Reason: "deployment must be an object"}
	}
	for _, req := range []string{"model", "profile_ref"} {
		if _, present := dep[req]; !present || string(dep[req]) == "null" {
			return &TargetEnvironmentError{Reason: "missing required deployment field: " + req}
		}
	}
	if len(dep) != 2 {
		return &TargetEnvironmentError{Reason: "unexpected field in deployment"}
	}

	var model string
	if json.Unmarshal(dep["model"], &model) != nil || len(model) < 1 {
		return &TargetEnvironmentError{Reason: "invalid deployment model"}
	}

	var profileRef string
	if json.Unmarshal(dep["profile_ref"], &profileRef) != nil || !validProfileRef(dep["profile_ref"], false) {
		return &TargetEnvironmentError{Reason: "invalid deployment profile_ref"}
	}
	return nil
}

func validateTargetWeb(raw json.RawMessage) error {
	if string(raw) == "null" {
		return nil
	}
	var web map[string]json.RawMessage
	if json.Unmarshal(raw, &web) != nil || web == nil {
		return &TargetEnvironmentError{Reason: "web must be an object"}
	}
	if _, present := web["server"]; !present || string(web["server"]) == "null" {
		return &TargetEnvironmentError{Reason: "web server is required"}
	}
	for k := range web {
		if k != "server" && k != "version" {
			return &TargetEnvironmentError{Reason: "unexpected field in web: " + k}
		}
	}
	var server string
	if json.Unmarshal(web["server"], &server) != nil || len(server) < 1 {
		return &TargetEnvironmentError{Reason: "invalid web server"}
	}
	if v, present := web["version"]; present && string(v) != "null" {
		var ver string
		if json.Unmarshal(v, &ver) != nil || len(ver) < 1 {
			return &TargetEnvironmentError{Reason: "invalid web version"}
		}
	}
	return nil
}

func validateTargetDatabase(raw json.RawMessage) error {
	if string(raw) == "null" {
		return nil
	}
	var db map[string]json.RawMessage
	if json.Unmarshal(raw, &db) != nil || db == nil {
		return &TargetEnvironmentError{Reason: "database must be an object"}
	}
	for _, req := range []string{"engine", "version", "preserve_schema"} {
		if _, present := db[req]; !present || string(db[req]) == "null" {
			return &TargetEnvironmentError{Reason: "missing required database field: " + req}
		}
	}
	if len(db) != 3 {
		return &TargetEnvironmentError{Reason: "unexpected field in database"}
	}
	var engine string
	if json.Unmarshal(db["engine"], &engine) != nil || len(engine) < 1 {
		return &TargetEnvironmentError{Reason: "invalid database engine"}
	}
	var ver string
	if json.Unmarshal(db["version"], &ver) != nil || len(ver) < 1 {
		return &TargetEnvironmentError{Reason: "invalid database version"}
	}
	var preserveSchema bool
	if json.Unmarshal(db["preserve_schema"], &preserveSchema) != nil || !preserveSchema {
		return &TargetEnvironmentError{Reason: "database preserve_schema must be true"}
	}
	return nil
}

func validateTargetCacheOrQueue(raw json.RawMessage, name string) error {
	if string(raw) == "null" {
		return nil
	}
	var obj map[string]json.RawMessage
	if json.Unmarshal(raw, &obj) != nil || obj == nil {
		return &TargetEnvironmentError{Reason: name + " must be an object"}
	}
	if _, present := obj["engine"]; !present || string(obj["engine"]) == "null" {
		return &TargetEnvironmentError{Reason: name + " engine is required"}
	}
	for k := range obj {
		if k != "engine" && k != "version" {
			return &TargetEnvironmentError{Reason: "unexpected field in " + name + ": " + k}
		}
	}
	var engine string
	if json.Unmarshal(obj["engine"], &engine) != nil || len(engine) < 1 {
		return &TargetEnvironmentError{Reason: "invalid " + name + " engine"}
	}
	if v, present := obj["version"]; present && string(v) != "null" {
		var ver string
		if json.Unmarshal(v, &ver) != nil || len(ver) < 1 {
			return &TargetEnvironmentError{Reason: "invalid " + name + " version"}
		}
	}
	return nil
}

func validateTargetStorage(raw json.RawMessage) error {
	if string(raw) == "null" {
		return nil
	}
	var st map[string]json.RawMessage
	if json.Unmarshal(raw, &st) != nil || st == nil {
		return &TargetEnvironmentError{Reason: "storage must be an object"}
	}
	for k := range st {
		if k != "local" && k != "engine" && k != "version" {
			return &TargetEnvironmentError{Reason: "unexpected field in storage: " + k}
		}
	}
	if loc, present := st["local"]; present && string(loc) != "null" {
		var local bool
		if json.Unmarshal(loc, &local) != nil {
			return &TargetEnvironmentError{Reason: "invalid storage local"}
		}
	}
	if eng, present := st["engine"]; present && string(eng) != "null" {
		var engine string
		if json.Unmarshal(eng, &engine) != nil || len(engine) < 1 {
			return &TargetEnvironmentError{Reason: "invalid storage engine"}
		}
	}
	if v, present := st["version"]; present && string(v) != "null" {
		var ver string
		if json.Unmarshal(v, &ver) != nil || len(ver) < 1 {
			return &TargetEnvironmentError{Reason: "invalid storage version"}
		}
	}
	return nil
}

func validateTargetPackageManager(raw json.RawMessage) error {
	if string(raw) == "null" {
		return nil
	}
	var pm map[string]json.RawMessage
	if json.Unmarshal(raw, &pm) != nil || pm == nil {
		return &TargetEnvironmentError{Reason: "package_manager must be an object"}
	}
	if _, present := pm["name"]; !present || string(pm["name"]) == "null" {
		return &TargetEnvironmentError{Reason: "package_manager name is required"}
	}
	for k := range pm {
		if k != "name" && k != "version" {
			return &TargetEnvironmentError{Reason: "unexpected field in package_manager: " + k}
		}
	}
	var name string
	if json.Unmarshal(pm["name"], &name) != nil || len(name) < 1 {
		return &TargetEnvironmentError{Reason: "invalid package_manager name"}
	}
	if v, present := pm["version"]; present && string(v) != "null" {
		var ver string
		if json.Unmarshal(v, &ver) != nil || len(ver) < 1 {
			return &TargetEnvironmentError{Reason: "invalid package_manager version"}
		}
	}
	return nil
}

func validateTargetCompatibility(raw json.RawMessage) error {
	if string(raw) == "null" {
		return nil
	}
	var comp map[string]json.RawMessage
	if json.Unmarshal(raw, &comp) != nil || comp == nil {
		return &TargetEnvironmentError{Reason: "compatibility must be an object"}
	}
	requiredCompat := []string{"http", "api", "auth", "database_behavior"}
	for _, req := range requiredCompat {
		if _, present := comp[req]; !present || string(comp[req]) == "null" {
			return &TargetEnvironmentError{Reason: "missing required compatibility field: " + req}
		}
		var mode string
		if json.Unmarshal(comp[req], &mode) != nil {
			return &TargetEnvironmentError{Reason: "invalid compatibility mode"}
		}
		if _, ok := allowedTargetCompatibilityModes[mode]; !ok {
			return &TargetEnvironmentError{Reason: "invalid compatibility mode enum"}
		}
	}
	for k := range comp {
		found := false
		for _, req := range requiredCompat {
			if k == req {
				found = true
				break
			}
		}
		if !found {
			return &TargetEnvironmentError{Reason: "unexpected field in compatibility: " + k}
		}
	}
	return nil
}

func parseTargetEnvironmentJSON(raw []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(raw) {
		return nil, &TargetEnvironmentError{Reason: "invalid UTF-8"}
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, &TargetEnvironmentError{Reason: "invalid JSON", Cause: err}
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return nil, &TargetEnvironmentError{Reason: "object fields"}
	}
	record := make(map[string]json.RawMessage)
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, &TargetEnvironmentError{Reason: "invalid object key", Cause: err}
		}
		key, ok := tok.(string)
		if !ok {
			return nil, &TargetEnvironmentError{Reason: "object key must be a string"}
		}
		if _, exists := record[key]; exists {
			return nil, &TargetEnvironmentError{Reason: "duplicate object key"}
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, &TargetEnvironmentError{Reason: "invalid object value", Cause: err}
		}
		record[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, &TargetEnvironmentError{Reason: "unterminated object", Cause: err}
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '}' {
		return nil, &TargetEnvironmentError{Reason: "unterminated object"}
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, &TargetEnvironmentError{Reason: "trailing input"}
	}
	return record, nil
}
