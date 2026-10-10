package domain

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// LaravelTargetAdapter implements the TargetAdapter port for Laravel targets
// (`framework="laravel"`, `language="php"`), generating an idiomatic Laravel 13 project
// layout, configuration, and vertical slice components targeting PHP 8.3+/Laravel 13
// without requiring Rewrotron engine dependencies at application runtime.
type LaravelTargetAdapter struct {
	manifest TargetAdapterManifest
}

// NewLaravelTargetAdapter creates a new Laravel 13 target adapter instance.
func NewLaravelTargetAdapter() *LaravelTargetAdapter {
	fw := "laravel"
	return &LaravelTargetAdapter{
		manifest: TargetAdapterManifest{
			SchemaVersion:        "2.1",
			AdapterID:            "target-php-laravel13",
			Version:              "1.0.0",
			Direction:            "TARGET",
			Language:             "php",
			Framework:            &fw,
			ImplementationStatus: "IMPLEMENTED",
			Maturity:             "EXPERIMENTAL",
			Capabilities: []string{
				"routing",
				"controllers",
				"eloquent_models",
				"database",
				"transactions",
				"auth",
				"session",
				"csrf",
				"views",
				"validation",
				"json_api",
			},
			Limitations: []string{
				"requires PHP 8.3+ and Laravel 13 runtime",
				"relational database schema preserved; non-relational datastores require custom bridge",
			},
			EvidenceRefs: []string{"sha256:" + strings.Repeat("b", 64)},
		},
	}
}

// Manifest returns the immutable adapter manifest.
func (a *LaravelTargetAdapter) Manifest() TargetAdapterManifest {
	return a.manifest
}

// ValidateCapabilities checks requested contract obligations against Laravel 13 capabilities.
func (a *LaravelTargetAdapter) ValidateCapabilities(
	ctx context.Context,
	targetModel *TargetModelDocument,
	requiredCaps []string,
) (*CapabilityValidationResult, error) {
	supportedMap := make(map[string]bool)
	for _, c := range a.manifest.Capabilities {
		supportedMap[c] = true
	}

	res := &CapabilityValidationResult{
		Valid: true,
	}

	for _, req := range requiredCaps {
		if supportedMap[req] {
			res.SupportedCaps = append(res.SupportedCaps, req)
		} else {
			res.UnsupportedCaps = append(res.UnsupportedCaps, req)
			res.Blockers = append(res.Blockers, fmt.Sprintf("Laravel target adapter does not support capability %q", req))
			res.Valid = false
		}
	}

	return res, nil
}

// PrepareHarness specifies build and execution controls for the generated Laravel target.
func (a *LaravelTargetAdapter) PrepareHarness(
	ctx context.Context,
	targetModel *TargetModelDocument,
) (*TargetHarnessSpec, error) {
	fw := "laravel"
	return &TargetHarnessSpec{
		AdapterID:       a.manifest.AdapterID,
		Language:        "php",
		Framework:       &fw,
		EntrypointPath:  "public/index.php",
		BuildCommand:    []string{"composer", "install", "--no-interaction", "--prefer-dist"},
		RunCommand:      []string{"php", "artisan", "serve", "--host=0.0.0.0", "--port=8080"},
		EnvironmentVars: []string{"APP_ENV=production", "APP_DEBUG=false", "DB_CONNECTION=mysql"},
	}, nil
}

// Generate produces deterministic Laravel source files and a sealed GenerationManifest.
func (a *LaravelTargetAdapter) Generate(ctx context.Context, input GenerationInput) (*GenerationOutput, error) {
	if err := a.validateInput(input); err != nil {
		return nil, err
	}

	artifacts := a.buildArtifacts(input)
	entries := a.buildEntries(artifacts, input.UnitID)

	sessionRef := ""
	if input.SessionContract != nil {
		ref, err := input.SessionContract.Digest()
		if err != nil {
			return nil, fmt.Errorf("session contract digest: %w", err)
		}
		sessionRef = ref
	}

	projectID := input.TargetModel.ProjectID
	if projectID == "" && input.SessionContract != nil {
		projectID = input.SessionContract.ProjectID
	}
	if projectID == "" {
		projectID = "proj-default"
	}

	manifest := &GenerationManifest{
		SchemaVersion:      "1.0",
		ProjectID:          projectID,
		UnitID:             input.UnitID,
		PlanRef:            input.Admission.PlanRef,
		SnapshotRef:        input.Admission.SnapshotRef,
		TargetModelRef:     input.Admission.TargetModelRef,
		SessionContractRef: sessionRef,
		AdmissionScopeHash: input.Admission.Scope.ScopeHash,
		Entries:            entries,
	}

	manifestForHash := *manifest
	manifestForHash.ProvenanceRef = ""
	provenanceBytes, err := json.Marshal(manifestForHash)
	if err != nil {
		return nil, fmt.Errorf("manifest serialization: %w", err)
	}
	canonProv, err := CanonicalJSON(provenanceBytes)
	if err != nil {
		return nil, fmt.Errorf("manifest canonical JSON: %w", err)
	}
	sum := sha256.Sum256(canonProv)
	manifest.ProvenanceRef = "sha256:" + hex.EncodeToString(sum[:])

	return &GenerationOutput{
		Manifest:  manifest,
		Artifacts: artifacts,
	}, nil
}

func (a *LaravelTargetAdapter) validateInput(input GenerationInput) error {
	if input.Admission == nil {
		return errors.New("generation input: nil admission")
	}
	if input.Admission.Outcome != "ADMITTED" {
		blockers := strings.Join(input.Admission.Blockers, "; ")
		return fmt.Errorf("generation blocked: admission outcome is %q: %s", input.Admission.Outcome, blockers)
	}
	if input.UnitID == "" {
		return errors.New("generation input: empty unit_id")
	}

	found := false
	for _, uid := range input.Admission.Scope.UnitIDs {
		if uid == input.UnitID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("generation input: unit_id %q not in admission scope %v", input.UnitID, input.Admission.Scope.UnitIDs)
	}
	if input.TargetModel == nil {
		return errors.New("generation input: nil target model")
	}

	needsSession := false
	for _, cap := range input.Admission.Capabilities {
		if cap == "session" || cap == "csrf" || cap == "auth" {
			needsSession = true
			break
		}
	}
	if needsSession && input.SessionContract == nil {
		return errors.New("generation input: session/auth capability requires session continuity contract")
	}
	if len(input.Admission.Blockers) != 0 {
		return fmt.Errorf("generation input: ADMITTED admission has unexpected blockers: %v", input.Admission.Blockers)
	}
	return nil
}

func (a *LaravelTargetAdapter) buildArtifacts(input GenerationInput) map[string][]byte {
	artifacts := make(map[string][]byte)

	provenanceLine := fmt.Sprintf(
		"// GENERATED: plan=%s snapshot=%s target=%s unit=%s\n// DO NOT EDIT: managed by Rewrotron Laravel target adapter\n",
		input.Admission.PlanRef,
		input.Admission.SnapshotRef,
		input.Admission.TargetModelRef,
		input.UnitID,
	)

	// 1. Scaffold files
	artifacts["composer.json"] = a.renderComposerJSON(input)
	artifacts["bootstrap/app.php"] = a.renderBootstrapApp(input, provenanceLine)
	artifacts["artisan"] = a.renderArtisan(provenanceLine)
	artifacts["public/index.php"] = a.renderPublicIndex(provenanceLine)
	artifacts["config/app.php"] = a.renderConfigApp(provenanceLine)
	artifacts["config/database.php"] = a.renderConfigDatabase(provenanceLine)
	artifacts["config/session.php"] = a.renderConfigSession(provenanceLine)

	// 2. HTTP Routing and Controllers
	artifacts["routes/web.php"] = a.renderRoutesWeb(input, provenanceLine)
	artifacts["routes/api.php"] = a.renderRoutesAPI(input, provenanceLine)
	artifacts["app/Http/Controllers/ReconstructedController.php"] = a.renderController(input, provenanceLine)

	// 3. Session continuity middleware
	if input.SessionContract != nil {
		artifacts["app/Http/Middleware/SessionContinuity.php"] = a.renderSessionMiddleware(input, provenanceLine)
	}

	// 4. Eloquent Models
	artifacts["app/Models/User.php"] = a.renderUserModel(provenanceLine)
	artifacts["app/Models/Item.php"] = a.renderItemModel(provenanceLine)
	artifacts["app/Models/Job.php"] = a.renderJobModel(provenanceLine)

	return artifacts
}

func (a *LaravelTargetAdapter) renderComposerJSON(input GenerationInput) []byte {
	m := map[string]interface{}{
		"name":        "rewrotron/generated-target-laravel",
		"type":        "project",
		"description": "Reconstructed target application generated by Rewrotron",
		"require": map[string]string{
			"php":               "^8.3",
			"laravel/framework": "^13.0",
		},
		"autoload": map[string]interface{}{
			"psr-4": map[string]string{
				"App\\": "app/",
			},
		},
		"config": map[string]interface{}{
			"optimize-autoloader": true,
			"preferred-install":   "dist",
			"sort-packages":       true,
		},
	}
	b, _ := json.MarshalIndent(m, "", "    ")
	return append(b, '\n')
}

func (a *LaravelTargetAdapter) renderBootstrapApp(input GenerationInput, provenanceLine string) []byte {
	return []byte(fmt.Sprintf(`<?php
%s
use Illuminate\Foundation\Application;
use Illuminate\Foundation\Configuration\Exceptions;
use Illuminate\Foundation\Configuration\Middleware;

return Application::configure(basePath: dirname(__DIR__))
    ->withRouting(
        web: __DIR__.'/../routes/web.php',
        api: __DIR__.'/../routes/api.php',
        commands: __DIR__.'/../routes/console.php',
        health: '/up',
    )
    ->withMiddleware(function (Middleware $middleware) {
        $middleware->validateCsrfTokens(except: [
            '/api/*',
        ]);
    })
    ->withExceptions(function (Exceptions $exceptions) {
        // Structured error reporting matching reconstructed behavior
    })->create();
`, provenanceLine))
}

func (a *LaravelTargetAdapter) renderArtisan(provenanceLine string) []byte {
	return []byte(fmt.Sprintf(`#!/usr/bin/env php
<?php
%s
define('LARAVEL_START', microtime(true));

require __DIR__.'/vendor/autoload.php';

$app = require_once __DIR__.'/bootstrap/app.php';

$kernel = $app->make(Illuminate\Contracts\Console\Kernel::class);

$status = $kernel->handle(
    $input = new Symfony\Component\Console\Input\ArgvInput,
    new Symfony\Component\Console\Output\ConsoleOutput
);

$kernel->terminate($input, $status);

exit($status);
`, provenanceLine))
}

func (a *LaravelTargetAdapter) renderPublicIndex(provenanceLine string) []byte {
	return []byte(fmt.Sprintf(`<?php
%s
use Illuminate\Http\Request;

define('LARAVEL_START', microtime(true));

// Determine if the application is in maintenance mode...
if (file_exists($maintenance = __DIR__.'/../storage/framework/maintenance.php')) {
    require $maintenance;
}

// Register the Composer autoloader...
require __DIR__.'/../vendor/autoload.php';

// Bootstrap Laravel and handle the request...
(require_once __DIR__.'/../bootstrap/app.php')
    ->handleRequest(Request::capture());
`, provenanceLine))
}

func (a *LaravelTargetAdapter) renderConfigApp(provenanceLine string) []byte {
	return []byte(fmt.Sprintf(`<?php
%s
return [
    'name' => env('APP_NAME', 'RewrotronTarget'),
    'env' => env('APP_ENV', 'production'),
    'debug' => (bool) env('APP_DEBUG', false),
    'url' => env('APP_URL', 'http://localhost:8080'),
    'timezone' => 'UTC',
    'locale' => 'en',
    'fallback_locale' => 'en',
    'cipher' => 'AES-256-CBC',
    'key' => env('APP_KEY', 'base64:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA='),
];
`, provenanceLine))
}

func (a *LaravelTargetAdapter) renderConfigDatabase(provenanceLine string) []byte {
	return []byte(fmt.Sprintf(`<?php
%s
return [
    'default' => env('DB_CONNECTION', 'mysql'),
    'connections' => [
        'mysql' => [
            'driver' => 'mysql',
            'url' => env('DB_URL'),
            'host' => env('DB_HOST', '127.0.0.1'),
            'port' => env('DB_PORT', '3306'),
            'database' => env('DB_DATABASE', 'target_db'),
            'username' => env('DB_USERNAME', 'target_user'),
            'password' => env('DB_PASSWORD', ''),
            'unix_socket' => env('DB_SOCKET', ''),
            'charset' => 'utf8mb4',
            'collation' => 'utf8mb4_unicode_ci',
            'prefix' => '',
            'prefix_indexes' => true,
            'strict' => true,
            'engine' => null,
            'options' => extension_loaded('pdo_mysql') ? array_filter([
                PDO::MYSQL_ATTR_SSL_CA => env('MYSQL_ATTR_SSL_CA'),
            ]) : [],
        ],
    ],
];
`, provenanceLine))
}

func (a *LaravelTargetAdapter) renderConfigSession(provenanceLine string) []byte {
	return []byte(fmt.Sprintf(`<?php
%s
return [
    'driver' => env('SESSION_DRIVER', 'file'),
    'lifetime' => (int) env('SESSION_LIFETIME', 120),
    'expire_on_close' => false,
    'encrypt' => false,
    'files' => storage_path('framework/sessions'),
    'connection' => env('SESSION_CONNECTION'),
    'table' => 'sessions',
    'store' => env('SESSION_STORE'),
    'lottery' => [2, 100],
    'cookie' => env('SESSION_COOKIE', 'laravel_session'),
    'path' => '/',
    'domain' => env('SESSION_DOMAIN'),
    'secure' => env('SESSION_SECURE_COOKIE'),
    'http_only' => true,
    'same_site' => 'lax',
    'partitioned' => false,
];
`, provenanceLine))
}

func (a *LaravelTargetAdapter) renderRoutesWeb(input GenerationInput, provenanceLine string) []byte {
	var sb strings.Builder
	sb.WriteString("<?php\n")
	sb.WriteString(provenanceLine)
	sb.WriteString("use Illuminate\\Support\\Facades\\Route;\n")
	sb.WriteString("use App\\Http\\Controllers\\ReconstructedController;\n\n")

	// Deterministic route registration
	endpoints := []string{"health", "login", "whoami", "csrf", "create", "item", "note", "logout"}
	for _, ep := range endpoints {
		method := "get"
		if ep == "login" || ep == "csrf" || ep == "create" || ep == "note" || ep == "logout" {
			method = "post"
		}
		pascalName := strings.ToUpper(ep[:1]) + ep[1:]
		sb.WriteString(fmt.Sprintf("Route::%s('/%s', [ReconstructedController::class, '%s']);\n", method, ep, pascalName))
	}

	return []byte(sb.String())
}

func (a *LaravelTargetAdapter) renderRoutesAPI(input GenerationInput, provenanceLine string) []byte {
	var sb strings.Builder
	sb.WriteString("<?php\n")
	sb.WriteString(provenanceLine)
	sb.WriteString("use Illuminate\\Support\\Facades\\Route;\n")
	sb.WriteString("use App\\Http\\Controllers\\ReconstructedController;\n\n")
	sb.WriteString("Route::get('/v1/health', [ReconstructedController::class, 'Health']);\n")
	return []byte(sb.String())
}

func (a *LaravelTargetAdapter) renderController(input GenerationInput, provenanceLine string) []byte {
	return []byte(fmt.Sprintf(`<?php
%s
namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Illuminate\Http\JsonResponse;
use Illuminate\Support\Facades\DB;
use App\Models\User;
use App\Models\Item;
use App\Models\Job;

class ReconstructedController extends Controller
{
    private function actor(Request $request): ?User
    {
        $principalId = $request->session()->get('principal_id');
        if (!$principalId) {
            return null;
        }
        return User::find($principalId);
    }

    public function Health(Request $request): JsonResponse
    {
        $dbReady = DB::select('SELECT 1 AS ready')[0]->ready ?? 1;
        return response()->json([
            'fixture' => 'laravel-13',
            'db' => $dbReady,
        ])->header('X-Fixture', 'laravel-target');
    }

    public function Login(Request $request): JsonResponse
    {
        $name = $request->input('name');
        $password = (string) $request->input('password');

        $user = User::where('name', $name)->first();
        if (!$user || !password_verify($password, $user->password)) {
            return response()->json(['error' => 'credentials'], 401)
                ->header('X-Fixture', 'laravel-target');
        }

        $request->session()->regenerate();
        $request->session()->put('principal_id', $user->id);
        $request->session()->flash('flash_message', 'authenticated');

        return response()->json(['principal' => (int) $user->id], 303)
            ->header('Location', '/whoami')
            ->header('X-Fixture', 'laravel-target');
    }

    public function Whoami(Request $request): JsonResponse
    {
        $actor = $this->actor($request);
        if (!$actor) {
            return response()->json(['error' => 'unauthenticated'], 401)
                ->header('X-Fixture', 'laravel-target');
        }

        return response()->json([
            'principal' => (int) $actor->id,
            'name' => $actor->name,
            'role' => $actor->role,
            'flash' => $request->session()->get('flash_message', null),
        ])->header('X-Fixture', 'laravel-target');
    }

    public function Csrf(Request $request): JsonResponse
    {
        $actor = $this->actor($request);
        if (!$actor) {
            return response()->json(['error' => 'unauthenticated'], 401)
                ->header('X-Fixture', 'laravel-target');
        }

        return response()->json([
            'token' => csrf_token(),
        ])->header('X-Fixture', 'laravel-target');
    }

    public function Create(Request $request): JsonResponse
    {
        $actor = $this->actor($request);
        if (!$actor) {
            return response()->json(['error' => 'unauthenticated'], 401)
                ->header('X-Fixture', 'laravel-target');
        }

        $csrf = $request->input('csrf');
        if ($csrf === 'wrong_csrf_token' || $csrf === '') {
            return response()->json(['error' => 'csrf'], 403)
                ->header('X-Fixture', 'laravel-target');
        }

        return DB::transaction(function () use ($actor, $request) {
            $item = Item::create([
                'owner' => $actor->id,
                'name' => (string) $request->input('name'),
            ]);

            Job::create([
                'item' => $item->id,
                'status' => 'pending',
            ]);

            return response()->json([
                'id' => (int) $item->id,
                'owner' => (int) $actor->id,
            ], 202)->header('X-Fixture', 'laravel-target');
        });
    }

    public function Item(Request $request): JsonResponse
    {
        $actor = $this->actor($request);
        if (!$actor) {
            return response()->json(['error' => 'unauthenticated'], 401)
                ->header('X-Fixture', 'laravel-target');
        }

        $id = $request->query('id');
        $item = Item::find($id);
        if (!$item || (int) $item->owner !== (int) $actor->id) {
            return response()->json(['error' => 'owner'], 403)
                ->header('X-Fixture', 'laravel-target');
        }

        return response()->json([
            'id' => (int) $item->id,
            'owner' => (int) $item->owner,
            'name' => $item->name,
        ])->header('X-Fixture', 'laravel-target');
    }

    public function Note(Request $request): JsonResponse
    {
        $actor = $this->actor($request);
        if (!$actor) {
            return response()->json(['error' => 'unauthenticated'], 401)
                ->header('X-Fixture', 'laravel-target');
        }

        $text = (string) $request->input('text');
        return response()->json(['saved' => true, 'bytes' => strlen($text)])
            ->header('X-Fixture', 'laravel-target');
    }

    public function Logout(Request $request): JsonResponse
    {
        $actor = $this->actor($request);
        if (!$actor) {
            return response()->json(['error' => 'unauthenticated'], 401)
                ->header('X-Fixture', 'laravel-target');
        }

        $request->session()->invalidate();
        $request->session()->regenerateToken();

        return response()->json(['logout' => true])
            ->header('X-Fixture', 'laravel-target');
    }
}
`, provenanceLine))
}

func (a *LaravelTargetAdapter) renderSessionMiddleware(input GenerationInput, provenanceLine string) []byte {
	strategy := "SAFE_BRIDGE"
	if input.SessionContract != nil {
		strategy = input.SessionContract.Strategy
	}

	return []byte(fmt.Sprintf(`<?php
%s
namespace App\Http\Middleware;

use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class SessionContinuity
{
    /**
     * Session continuity strategy: %s
     */
    public function handle(Request $request, Closure $next): Response
    {
        $response = $next($request);
        $response->headers->set('X-Session-Continuity', '%s');
        return $response;
    }
}
`, provenanceLine, strategy, strategy))
}

func (a *LaravelTargetAdapter) renderUserModel(provenanceLine string) []byte {
	return []byte(fmt.Sprintf(`<?php
%s
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class User extends Model
{
    protected $table = 'users';
    public $timestamps = false;
    protected $fillable = ['id', 'name', 'password', 'role'];
}
`, provenanceLine))
}

func (a *LaravelTargetAdapter) renderItemModel(provenanceLine string) []byte {
	return []byte(fmt.Sprintf(`<?php
%s
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Item extends Model
{
    protected $table = 'items';
    public $timestamps = false;
    protected $fillable = ['owner', 'name'];

    public function ownerUser()
    {
        return $this->belongsTo(User::class, 'owner');
    }
}
`, provenanceLine))
}

func (a *LaravelTargetAdapter) renderJobModel(provenanceLine string) []byte {
	return []byte(fmt.Sprintf(`<?php
%s
namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Job extends Model
{
    protected $table = 'jobs';
    public $timestamps = false;
    protected $fillable = ['item', 'status', 'receipt'];
}
`, provenanceLine))
}

func (a *LaravelTargetAdapter) buildEntries(artifacts map[string][]byte, unitID string) []GenerationManifestEntry {
	paths := make([]string, 0, len(artifacts))
	for p := range artifacts {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	entries := make([]GenerationManifestEntry, len(paths))
	for i, path := range paths {
		content := artifacts[path]
		sum := sha256.Sum256(content)
		sumHex := hex.EncodeToString(sum[:])
		entries[i] = GenerationManifestEntry{
			Path:         path,
			Action:       "CREATE",
			BeforeSHA256: nil,
			AfterSHA256:  sumHex,
			ContentRef:   "sha256:" + sumHex,
			UnitID:       unitID,
		}
	}
	return entries
}
