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

        // Get the user ID (from session or other auth system)
        $userId = $this->getAuthenticatedUserId();
        
        // If the user is authenticated, set the user ID in Aikido Zen context
        if ($userId) {
            \aikido\set_user($userId);
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
            else if ($decision->trigger == "ip") {
                $message = "Your IP ({$decision->ip}) is blocked due to: {$decision->description}!";
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
}
```

## Laravel

```php
<?php

namespace App\Http\Middleware;

use Closure;
use Illuminate\Support\Facades\Auth;

class ZenBlockDecision
{
    public function handle($request, Closure $next)
    {
        // Check if Aikido extension is loaded
        if (!extension_loaded('aikido')) {
            return $next($request);
        }
		
		// Get the authenticated user's ID from Laravel's Auth system
		$userId = Auth::id();

		// If a user is authenticated, set the user in Aikido's firewall context
		if ($userId) {
			\aikido\set_user($userId);
		}

        // Check blocking decision from Aikido
        $decision = \aikido\should_block_request();

        if ($decision->block) {
            if ($decision->type == "blocked") {
                if ($decision->trigger == "user") {
                    return response('Your user is blocked!', 403);
                }
                else if ($decision->trigger == "ip") {
                    return response("Your IP ({$decision->ip}) is blocked due to: {$decision->description}!", 403);
                }
            }
            else if ($decision->type == "ratelimited") {
                if ($decision->trigger == "user") {
                    return response('Your user exceeded the rate limit for this endpoint!', 429);
                }
                else if ($decision->trigger == "ip") {
                    return response("Your IP ({$decision->ip}) exceeded the rate limit for this endpoint!", 429);
                }
            }
        }

        // Continue to the next middleware or request handler
        return $next($request);
    }
}
```
