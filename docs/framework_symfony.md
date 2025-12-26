---
title: Symfony
eleventyNavigation:
  key: Symfony
  parent: Frameworks
showNavChildren: true
---

# Zen Firewall For Symfony

{% renderFile "./shared-intro.md" %}

## Install the Agent

{% renderFile "./shared-package.md" %}

## Configure the environment variable

{% renderFile "./shared-environment-variable.md" %}

## Install the Middleware

### 1. Place the AikidoMiddleware in your app:

```php
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

### 2. Enable Middleware

...

## Troubleshooting

{% renderFile "./shared-troubleshooting.md" %}
