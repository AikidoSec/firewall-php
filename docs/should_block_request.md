# Should block request

In order to enable the user blocking and rate limiting features, the protected app can call `\aikido\should_block_request` to obtain the blocking decision for the current request and act accordingly.
We provide middleware examples that can be used in different scenarious.

## No framework

```php
<?php

namespace App\Middleware;

use Psr\Http\Message\ResponseInterface;
use Psr\Http\Message\ServerRequestInterface;
use Psr\Http\Server\MiddlewareInterface;
use Psr\Http\Server\RequestHandlerInterface;
use Laminas\Diactoros\Response; // Or use any other PSR-7 implementation

class AikidoMiddleware implements MiddlewareInterface
{
    public function process(ServerRequestInterface $request, RequestHandlerInterface $handler): ResponseInterface
    {
        // Start the session (if needed) to track user login status
        session_start();

        // Check if Aikido extension is loaded
        if (!extension_loaded('aikido')) {
            // Extension not loaded
            // Pass the request to the next middleware or request handler
            return $handler->handle($request);
        }

        // You can pass in the Aikido token here
        // \aikido\set_token("your token here");

        // Get the user ID / name (from session or other auth system)
        $userId = $this->getAuthenticatedUserId();

        // If the user is authenticated, set the user ID in Aikido Zen context
        if ($userId) {
            // Username is optional: \aikido\set_user can be called only with user ID
            $userName = $this->getAuthenticatedUserName();
            \aikido\set_user($userId, $userName);
        }

        // Check blocking decision from Aikido
        $decision = \aikido\should_block_request();

        if (!$decision->block) {
            // Aikido decided not to block
            // Pass the request to the next middleware or request handler
            return $handler->handle($request);
        }

        if ($decision->type == "blocked") {
            // If the user/ip is blocked, return a 403 status code
            $message = "";
            if ($decision->trigger == "user") {
                $message = "Your user is blocked!";
            }

            return new Response([
                'message' => $message,
            ], 403);
        }
        else if ($decision->type == "ratelimited") {
            // If the rate limit is exceeded, return a 429 status code
            $message = "";
            if ($decision->trigger == "user") {
                $message = "Your user exceeded the rate limit for this endpoint!";
            }
            else if ($decision->trigger == "ip") {
                $message = "Your IP ({$decision->ip}) exceeded the rate limit for this endpoint!";
            }
            else if ($decision->trigger == "group") {
                $message = "Your group exceeded the rate limit for this endpoint!";
            }
            return new Response([
                'message' => $message,
            ], 429);
        }

        // Aikido decided to block but decision type is not implemented
        return new Response([
            'message' => 'Something went wrong!',
        ], 500);
    }

    // Example function to simulate user authentication
    private function getAuthenticatedUserId(): ?int
    {
        return $_SESSION['user_id'] ?? null;
    }
    // Example function to simulate user authentication
    private function getAuthenticatedUserName(): ?string
    {
        return $_SESSION['user_name'] ?? null;
    }
}
```

## Laravel

1. Place the AikidoMiddleware in `app/Http/Middleware/AikidoMiddleware.php`.

```php
<?php

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

2. In `bootstrap/app.php`, apply the following changes:
```php
<?php
// ...

use App\Http\Middleware\AikidoMiddleware;

return Application::configure(basePath: dirname(__DIR__))
    ->withRouting(
        web: __DIR__.'/../routes/web.php',
        commands: __DIR__.'/../routes/console.php',
        health: '/up',
    )
    ->withMiddleware(function (Middleware $middleware): void {
        $middleware->web(append: [
            AikidoMiddleware::class,
        ]);
        // Append AikidoMiddleware to other groups ('api' for example)
    })

// ...
```