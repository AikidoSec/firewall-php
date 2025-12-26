---
title: Laravel
eleventyNavigation:
  key: Laravel
  parent: Frameworks
showNavChildren: true
---

# Zen Firewall For Laravel

{% renderFile "./shared-intro.md" %}

## Install the Agent

{% renderFile "./shared-package.md" %}

## Configure the environment variable

{% renderFile "./shared-environment-variable.md" %}

## Install the Middleware

### 1. Place the AikidoMiddleware in `app/Http/Middleware/AikidoMiddleware.php`:

```php
namespace App\Http\Middleware;

use Closure;
use Illuminate\Support\Facades\Auth;

class AikidoMiddleware
{
    public function handle($request, Closure $next)
    {
        // Check if Aikido extension is loaded
        if (!extension_loaded('aikido')) {
            return $next($request);
        }

        // You can pass in the Aikido token here
        // \aikido\set_token("your token here");

		
        // Get the authenticated user's ID from Laravel's Auth system
        $userId = Auth::id();

        // If a user is authenticated, set the user in Aikido Zen context
        if ($userId) {
            \aikido\set_user($userId);
            // If you want to set the user's name in Aikido Zen context, you can change the above to:
            // \aikido\set_user($userId, Auth::user()?->name);
        }

        // Check blocking decision from Aikido
        $decision = \aikido\should_block_request();

        if ($decision->block) {
            if ($decision->type == "blocked") {
                if ($decision->trigger == "user") {
                    return response('Your user is blocked!', 403);
                }
            }
            else if ($decision->type == "ratelimited") {
                if ($decision->trigger == "user") {
                    return response('Your user exceeded the rate limit for this endpoint!', 429);
                }
                else if ($decision->trigger == "ip") {
                    return response("Your IP ({$decision->ip}) exceeded the rate limit for this endpoint!", 429);
                }
                else if ($decision->trigger == "group") {
                    return response("Your group exceeded the rate limit for this endpoint!", 429);
                }
            }
        }

        // Continue to the next middleware or request handler
        return $next($request);
    }
}
```

### 2. Enable Middleware

In `bootstrap/app.php`, apply the following changes:
```diff-php
use App\Http\Middleware\AikidoMiddleware;

return Application::configure(basePath: dirname(__DIR__))
    ->withRouting(
        web: __DIR__.'/../routes/web.php',
        commands: __DIR__.'/../routes/console.php',
        health: '/up',
    )
    ->withMiddleware(function (Middleware $middleware): void {
        $middleware->web(append: [
+            AikidoMiddleware::class,
        ]);
+        // Append AikidoMiddleware to other groups ('api' for example)
    })
```

## Troubleshooting

{% renderFile "./shared-troubleshooting.md" %}
