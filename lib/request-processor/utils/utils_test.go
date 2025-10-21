package utils

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	. "main/aikido_types"
	"main/globals"
	"math/big"
	"regexp"
	"strings"
	"testing"
)

var (
	lower    = "abcdefghijklmnopqrstuvwxyz"
	upper    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers  = "0123456789"
	specials = "!#$%^&*|;:<>"
)

func secretFromCharset(length int, charset string) string {
	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

func TestLooksLikeASecret(t *testing.T) {
	t.Run("it returns false for empty string", func(t *testing.T) {
		if LooksLikeASecret("") {
			t.Errorf("expected false for empty string")
		}
	})

	t.Run("it returns false for short strings", func(t *testing.T) {
		shortStrings := []string{"c", "NR", "7t3", "4qEK", "KJr6s", "KXiW4a", "Fupm2Vi", "jiGmyGfg", "SJPLzVQ8t", "OmNf04j6mU"}
		for _, s := range shortStrings {
			if LooksLikeASecret(s) {
				t.Errorf("expected false for short string %s", s)
			}
		}
	})

	t.Run("it returns true for long strings", func(t *testing.T) {
		longStrings := []string{"rsVEExrR2sVDONyeWwND", ":2fbg;:qf$BRBc<2AG8&"}
		for _, s := range longStrings {
			if !LooksLikeASecret(s) {
				t.Errorf("expected true for long string %s", s)
			}
		}
	})

	t.Run("it flags very long strings", func(t *testing.T) {
		veryLongString := "efDJHhzvkytpXoMkFUgag6shWJktYZ5QUrUCTfecFELpdvaoAT3tekI4ZhpzbqLt"
		if !LooksLikeASecret(veryLongString) {
			t.Errorf("expected true for very long string")
		}
	})

	t.Run("it flags very very long strings", func(t *testing.T) {
		veryVeryLongString := "XqSwF6ySwMdTomIdmgFWcMVXWf5L0oVvO5sIjaCPI7EjiPvRZhZGWx3A6mLl1HXPOHdUeabsjhngW06JiLhAchFwgtUaAYXLolZn75WsJVKHxEM1mEXhlmZepLCGwRAM"
		if !LooksLikeASecret(veryVeryLongString) {
			t.Errorf("expected true for very very long string")
		}
	})

	t.Run("it returns false if contains white space", func(t *testing.T) {
		if LooksLikeASecret("rsVEExrR2sVDONyeWwND ") {
			t.Errorf("expected false for string with white space")
		}
	})

	t.Run("it returns false if it has less than 2 charsets", func(t *testing.T) {
		if LooksLikeASecret(secretFromCharset(10, lower)) {
			t.Errorf("expected false for string with only lower case letters")
		}
		if LooksLikeASecret(secretFromCharset(10, upper)) {
			t.Errorf("expected false for string with only upper case letters")
		}
		if LooksLikeASecret(secretFromCharset(10, numbers)) {
			t.Errorf("expected false for string with only numbers")
		}
		if LooksLikeASecret(secretFromCharset(10, specials)) {
			t.Errorf("expected false for string with only special characters")
		}
	})

	urlTerms := []string{
		"development", "programming", "applications", "implementation", "environment", "technologies",
		"documentation", "demonstration", "configuration", "administrator", "visualization",
		"international", "collaboration", "opportunities", "functionality", "customization",
		"specifications", "optimization", "contributions", "accessibility", "subscription",
		"subscriptions", "infrastructure", "architecture", "authentication", "sustainability",
		"notifications", "announcements", "recommendations", "communication", "compatibility",
		"enhancement", "integration", "performance", "improvements", "introduction", "capabilities",
		"communities", "credentials", "integration", "permissions", "validation", "serialization",
		"deserialization", "rate-limiting", "throttling", "load-balancer", "microservices",
		"endpoints", "data-transfer", "encryption", "authorization", "bearer-token", "multipart",
		"urlencoded", "api-docs", "postman", "json-schema", "serialization", "deserialization",
		"rate-limiting", "throttling", "load-balancer", "api-gateway", "microservices", "endpoints",
		"data-transfer", "encryption", "signature", "poppins-bold-webfont.woff2", "karla-bold-webfont.woff2",
		"startEmailBasedLogin", "jenkinsFile", "ConnectionStrings.config", "coach", "login", "payment_methods",
		"activity_logs", "feedback_responses", "balance_transactions", "customer_sessions", "payment_intents",
		"billing_portal", "subscription_items", "namedLayouts", "PlatformAction", "quickActions", "queryLocator",
		"relevantItems", "parameterizedSearch",
	}

	t.Run("it returns false for common url terms", func(t *testing.T) {
		for _, term := range urlTerms {
			if LooksLikeASecret(term) {
				t.Errorf("expected false for common url term %s", term)
			}
		}
	})

	t.Run("it returns false for known word separators", func(t *testing.T) {
		if LooksLikeASecret("this-is-a-secret-1") {
			t.Errorf("expected false for string with known word separators")
		}
	})

	t.Run("a number is not a secret", func(t *testing.T) {
		if LooksLikeASecret("1234567890") {
			t.Errorf("expected false for number string 1234567890")
		}
		if LooksLikeASecret("12345678901234567890") {
			t.Errorf("expected false for number string 12345678901234567890")
		}
	})

	secrets := []string{
		"yqHYTS<agpi^aa1",
		"hIofuWBifkJI5iVsSNKKKDpBfmMqJJwuXMxau6AS8WZaHVLDAMeJXo3BwsFyrIIm",
		"AG7DrGi3pDDIUU1PrEsj",
		"CnJ4DunhYfv2db6T1FRfciRBHtlNKOYrjoz",
		"Gic*EfMq:^MQ|ZcmX:yW1",
		"AG7DrGi3pDDIUU1PrEsj",
	}

	t.Run("it returns true for known secrets", func(t *testing.T) {
		for _, secret := range secrets {
			if !LooksLikeASecret(secret) {
				t.Errorf("expected true for known secret %s", secret)
			}
		}
	})
}

func generateHash(algorithm string) string {
	data := []byte("test")

	switch algorithm {
	case "md5":
		hash := md5.Sum(data)
		return hex.EncodeToString(hash[:])
	case "sha1":
		hash := sha1.Sum(data)
		return hex.EncodeToString(hash[:])
	case "sha256":
		hash := sha256.Sum256(data)
		return hex.EncodeToString(hash[:])
	case "sha512":
		hash := sha512.Sum512(data)
		return hex.EncodeToString(hash[:])
	default:
		return ""
	}
}

func TestBuildRouteFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"", ""},
		{"http", "http"},
		{"/", "/"},
		{"http://localhost/", "/"},
		{"/posts/3", "/posts/:number"},
		{"http://localhost/posts/3", "/posts/:number"},
		{"http://localhost/posts/3/", "/posts/:number"},
		{"http://localhost/posts/3/comments/10", "/posts/:number/comments/:number"},
		{"/blog/2023/05/great-article", "/blog/:number/:number/great-article"},
		{"/posts/2023-05-01", "/posts/:date"},
		{"/posts/2023-05-01/", "/posts/:date"},
		{"/posts/2023-05-01/comments/2023-05-01", "/posts/:date/comments/:date"},
		{"/posts/01-05-2023", "/posts/:date"},
		{"/posts/3,000", "/posts/3,000"},
		{"/v1/posts/3", "/v1/posts/:number"},
		{"/posts/d9428888-122b-11e1-b85c-61cd3cbb3210", "/posts/:uuid"},
		{"/posts/000003e8-2363-21ef-b200-325096b39f47", "/posts/:uuid"},
		{"/posts/a981a0c2-68b1-35dc-bcfc-296e52ab01ec", "/posts/:uuid"},
		{"/posts/109156be-c4fb-41ea-b1b4-efe1671c5836", "/posts/:uuid"},
		{"/posts/90123e1c-7512-523e-bb28-76fab9f2f73d", "/posts/:uuid"},
		{"/posts/1ef21d2f-1207-6660-8c4f-419efbd44d48", "/posts/:uuid"},
		{"/posts/017f22e2-79b0-7cc3-98c4-dc0c0c07398f", "/posts/:uuid"},
		{"/posts/0d8f23a0-697f-83ae-802e-48f3756dd581", "/posts/:uuid"},
		{"/posts/ECBCDD2C-A441-4846-B5AC-0083D347FDF2", "/posts/:uuid"},
		{"/posts/00000000-0000-1000-6000-000000000000", "/posts/00000000-0000-1000-6000-000000000000"},
		{"/posts/abc", "/posts/abc"},
		{"/login/john.doe@acme.com", "/login/:email"},
		{"/login/john.doe+alias@acme.com", "/login/:email"},
		{"/block/1.2.3.4", "/block/:ip"},
		{"/block/2001:2:ffff:ffff:ffff:ffff:ffff:ffff", "/block/:ip"},
		{"/block/64:ff9a::255.255.255.255", "/block/:ip"},
		{"/block/100::", "/block/:ip"},
		{"/block/fec0::", "/block/:ip"},
		{"/block/227.202.96.196", "/block/:ip"},
		{"/files/" + generateHash("md5"), "/files/:hash"},
		{"/files/" + generateHash("sha1"), "/files/:hash"},
		{"/files/" + generateHash("sha256"), "/files/:hash"},
		{"/files/" + generateHash("sha512"), "/files/:hash"},
		{"/confirm/CnJ4DunhYfv2db6T1FRfciRBHtlNKOYrjoz", "/confirm/:secret"},
		{"/posts/01ARZ3NDEKTSV4RRFFQ69G5FAV", "/posts/:ulid"},
		{"/posts/01arz3ndektsv4rrffq69g5fav", "/posts/:ulid"},
		{"/posts/66ec29159d00113616fc7184", "/posts/:objectId"},
		{"/posts/66EC29159D00113616FC7184", "/posts/:objectId"},
		{"/files/" + strings.ToUpper(generateHash("md5")), "/files/:hash"},
		{"/files/" + strings.ToUpper(generateHash("sha1")), "/files/:hash"},
		{"/files/" + strings.ToUpper(generateHash("sha256")), "/files/:hash"},
		{"/files/" + strings.ToUpper(generateHash("sha512")), "/files/:hash"},
	}

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			result := BuildRouteFromURL(test.url)
			if result != test.expected {
				t.Errorf("expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestParseBodyJSON(t *testing.T) {
	data := "\r\n\r\n\r\n{\r\n\r\n\r\n\"a\":\r\n\r\n\r\n \"1\",\r\n\r\n\"b\":\"2\"\r\n\r\n}\r\n\r\n\r\n\r\n"
	expected := `{"a":"1","b":"2"}`

	result := ParseBody(data)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		t.Errorf("Failed to marshal result: %v", err)
	}

	if string(resultJSON) != expected {
		t.Errorf("Expected JSON string %q, got %q", expected, resultJSON)
	}
	data = "{ \"age\": -1e+9999, \"cmd\": \"cat /etc/passwd\"}"
	expected = `{"age":-1e+9999,"cmd":"cat /etc/passwd"}`

	result = ParseBody(data)
	resultJSON, err = json.Marshal(result)
	if err != nil {
		t.Errorf("Failed to marshal result: %v", err)
	}

	if string(resultJSON) != expected {
		t.Errorf("Expected JSON string %q, got %q", expected, resultJSON)
	}
}

func TestParseBodyJSONArray(t *testing.T) {
	data := `["asd",  "asd"]`
	expected := `{"array":["asd","asd"]}`

	result := ParseBody(data)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		t.Errorf("Failed to marshal result: %v", err)
	}

	if string(resultJSON) != expected {
		t.Errorf("Expected JSON string %q, got %q", expected, string(resultJSON))
	}
}

func TestParseCookie(t *testing.T) {
	data := "exploit=/etc/passwd;exploit=safevalue;"
	result := ParseFormData(data, ";")
	if result["exploit"] != "/etc/passwd" {
		t.Errorf("Expected /etc/passwd, got %v", result["exploit"])
	}
}

func TestParseFormData(t *testing.T) {
	data := "a=1&b=2"
	result := ParseFormData(data, "&")
	if result["a"] != "1" {
		t.Errorf("Expected 1, got %v", result["a"])
	}
	if result["b"] != "2" {
		t.Errorf("Expected 2, got %v", result["b"])
	}

	data = "id=1+AND+sleep(3)--+="
	result = ParseFormData(data, "&")
	if result["id"] != "1 AND sleep(3)-- =" {
		t.Errorf("Expected 1 AND sleep(3)-- =, got %v", result["id"])
	}
}

func TestParseFormDataWithInvalidEncoding(t *testing.T) {
	data := "a=1asdlasdasd%22eee%ZZAAA&b=%20example%ZZPADDINGfoo%20bar%ZZ"
	result := ParseFormData(data, "&")
	if result["b"] != " example%ZZPADDINGfoo bar%ZZ" {
		t.Errorf("retrieved invalid url decoded value: %v", result["b"])
	}
}
func TestDecodeURIComponent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic string without encoding",
			input:    "a=1&b=2",
			expected: "a=1&b=2",
		},
		{
			name:     "basic percent encoding",
			input:    "hello%20world",
			expected: "hello world",
		},
		{
			name:     "plus sign encoding",
			input:    "hello+world",
			expected: "hello world",
		},
		{
			name:     "mixed encoding",
			input:    "hello%20world+and%20universe",
			expected: "hello world and universe",
		},
		{
			name:     "special characters",
			input:    "%21%40%23%24%25%5E%26%2A%28%29",
			expected: "!@#$%^&*()",
		},
		{
			name:     "invalid percent encoding - incomplete",
			input:    "test%2",
			expected: "test%2",
		},
		{
			name:     "invalid percent encoding - invalid hex",
			input:    "test%ZZ",
			expected: "test%ZZ",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "percent sign without encoding",
			input:    "100% complete",
			expected: "100% complete",
		},
		{
			name:     "multiple consecutive percent encodings",
			input:    "%20%20%20",
			expected: "   ",
		},
		{
			name:     "mixed case hex encoding",
			input:    "%2f%2F%3a%3A",
			expected: "//::"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DecodeURIComponent(tt.input)
			if result != tt.expected {
				t.Errorf("DecodeURIComponent(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsUserAgentBlocked(t *testing.T) {
	pattern := "Applebot-Extended|archive.org_bot|Arquivo-web-crawler|heritrix|ia_archiver|NiceCrawler|AhrefsBot|AhrefsSiteAudit|Barkrowler|BLEXBot|BrightEdge Crawler|Cocolyzebot|DataForSeoBot|DomainStatsBot|dotbot|hypestat|linkdexbot|MJ12bot|online-webceo-bot|Screaming Frog SEO Spider|SemrushBot|SenutoBot|SeobilityBot|SEOkicks|SEOlizer|serpstatbot|SiteCheckerBotCrawler|SenutoBot|ZoomBot|Seodiver|SEOlyzer|Backlinkcrawler|rogerbot|Siteimprove\\.com|360Spider|AlexandriaOrgBot|Baiduspider|bingbot|coccocbot-web|Daum|DuckDuckBot|DuckDuckGo-Favicons-Bot|Feedfetcher-Google|Google Favicon|Googlebot|GoogleOther|HaoSouSpider|MojeekBot|msnbot|PetalBot|Qwantbot|Qwantify|SemanticScholarBot|SeznamBot|Sogou web spider|teoma|TinEye|yacybot|Yahoo! Slurp|Yandex|Yeti|YisouSpider|ZumBot|AntBot|Amazonbot|Applebot|OAI-SearchBot|PerplexityBot|YouBot|sqlmap|WPScan|feroxbuster|masscan|Fuzz Faster U Fool|gobuster|\\(hydra\\)|absinthe|arachni|bsqlbf|cisco-torch|crimscanner|DirBuster|Grendel-Scan|Mysqloit|Nmap NSE|Nmap Scripting Engine|Nessus|Netsparker|Nikto|Paros|uil2pn|SQL Power Injector|webshag|Teh Forest Lobster|DotDotPwn|Havij|OpenVAS|ZmEu|DominoHunter|domino hunter|FHScan Core|w3af\\.(sf\\.net|sourceforge\\.net|org)|cgichk|webvulnscan|sqlninja|Argus(-Scanner|Crawler|DataLeakChecker|Bot)|ShadowSpray\\.Kerb|OWASP Amass|Argus(-Scanner|Crawler|DataLeakChecker|Bot)|Nuclei|BackDoorBot|HeadlessChrome|HeadlessEdg|facebookexternalhit|facebookcatalog|meta-externalagent|meta-externalfetcher|Twitterbot|Pinterestbot|pinterest\\.com.bot|LinkedInBot|XING-contenttabreceiver|redditbot|Mastodon|Bluesky Cardyb|vkShare|EmailCollector|EmailSiphon|EmailWolf|ExtractorPro|MailSweeper|Email Extractor|WebDataExtractor|MailBait"

	server := globals.NewServerData()
	server.CloudConfig.BlockedUserAgents = regexp.MustCompile(pattern)

	tests := []struct {
		ua       string
		expected bool
	}{
		{"Googlebot", true},
		{"AhrefsBot", true},
		{"SemrushBot/7.0", true},
		{"Mozilla/5.0 (compatible; Bingbot/2.0; +http://www.bing.com/bingbot.htm)", true},
		{"facebookexternalhit/1.1", true},
		{"LinkedInBot/1.0", true},
		{"Twitterbot/1.0", true},
		{"HeadlessChrome", true},
		{"Nuclei", true},
		{"DotDotPwn", true},
		{"sqlmap", true},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", false},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", false},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/537.36 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/537.36", false},
		{"Mozilla/5.0 (Linux; Android 10; SM-G973F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Mobile Safari/537.36", false},
		{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", false},
	}

	for _, test := range tests {
		t.Run(test.ua, func(t *testing.T) {
			result, _ := IsUserAgentBlocked(server, test.ua)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestIsIpBlockedByPrefix(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"1.2.0.0/16"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "1.2.3.4"
	result, _ := IsIpBlocked(server, ip)
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestIsIpBlockedByIp(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"1.2.3.4"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "1.2.3.4"
	result, _ := IsIpBlocked(server, ip)
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestIsIpNotBlockedByPrefix(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"1.2.0.0/16"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "2.3.4.5"
	result, _ := IsIpBlocked(server, ip)
	if result != false {
		t.Errorf("expected false, got %v", result)
	}
}

func TestIsIpNotBlockedByIp(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"1.2.3.4"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "2.3.4.5"
	result, _ := IsIpBlocked(server, ip)
	if result != false {
		t.Errorf("expected false, got %v", result)
	}
}
func TestIsIpv6BlockedByPrefix(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"2001:db8::/32"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "2001:db8:1234:5678:90ab:cdef:1234:5678"
	result, _ := IsIpBlocked(server, ip)
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestIsIpv6BlockedByIp(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"2001:db8::1"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "2001:db8::1"
	result, _ := IsIpBlocked(server, ip)
	if result != true {
		t.Errorf("expected true, got %v", result)
	}
}

func TestIsIpv6NotBlockedByPrefix(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"2001:db8::/32"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "2001:db9::1"
	result, _ := IsIpBlocked(server, ip)
	if result != false {
		t.Errorf("expected false, got %v", result)
	}
}

func TestIsIpv6NotBlockedByIp(t *testing.T) {
	server := globals.NewServerData()
	server.CloudConfig.BlockedIps = map[string]IpList{}
	IpList, _ := BuildIpList("test", []string{"2001:db8::1"})
	server.CloudConfig.BlockedIps["test"] = *IpList
	ip := "2001:db8::2"
	result, _ := IsIpBlocked(server, ip)
	if result != false {
		t.Errorf("expected false, got %v", result)
	}
}

func TestGetIpFromRequest(t *testing.T) {
	//no headers and no remote address
	server := globals.NewServerData()
	server.AikidoConfig.TrustProxy = false
	if got := GetIpFromRequest(server, "", ""); got != "" {
		t.Errorf("expected empty, got %q", got)
	}

	server.AikidoConfig.TrustProxy = true
	if got := GetIpFromRequest(server, "", ""); got != "" {
		t.Errorf("expected empty, got %q", got)
	}

	//no headers and remote address
	server.AikidoConfig.TrustProxy = false
	if got := GetIpFromRequest(server, "1.2.3.4", ""); got != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %q", got)
	}

	server.AikidoConfig.TrustProxy = true
	if got := GetIpFromRequest(server, "1.2.3.4", ""); got != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %q", got)
	}

	// x-forwarded-for without trust proxy
	server.AikidoConfig.TrustProxy = false
	if got := GetIpFromRequest(server, "1.2.3.4", "9.9.9.9"); got != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %q", got)
	}

	if got := GetIpFromRequest(server, "df89:84af:85e0:c55f:960c:341a:2cc6:734d", "a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880"); got != "df89:84af:85e0:c55f:960c:341a:2cc6:734d" {
		t.Errorf("expected df89:84af:85e0:c55f:960c:341a:2cc6:734d, got %q", got)
	}

	// x-forwarded-for with trust proxy and "x-forwarded-for" is not an IP
	server.AikidoConfig.TrustProxy = true
	if got := GetIpFromRequest(server, "1.2.3.4", "invalid"); got != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %q", got)
	}

	// x-forwarded-for with trust proxy and IP contains port
	server.AikidoConfig.TrustProxy = true
	if got := GetIpFromRequest(server, "1.2.3.4", "9.9.9.9:8080"); got != "9.9.9.9" {
		t.Errorf("expected 9.9.9.9, got %q", got)
	}
	if got := GetIpFromRequest(server, "1.2.3.4", "[a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880]:8080"); got != "a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880" {
		t.Errorf("expected a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880, got %q", got)
	}
	if got := GetIpFromRequest(server, "1.2.3.4", "[a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880]"); got != "a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880" {
		t.Errorf("expected a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880, got %q", got)
	}
	// Invalid format
	if got := GetIpFromRequest(server, "df89:84af:85e0:c55f:960c:341a:2cc6:734d", "a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880:8080"); got != "df89:84af:85e0:c55f:960c:341a:2cc6:734d" {
		t.Errorf("expected df89:84af:85e0:c55f:960c:341a:2cc6:734d, got %q", got)
	}

	// with trailing comma
	server.AikidoConfig.TrustProxy = true
	if got := GetIpFromRequest(server, "1.2.3.4", "9.9.9.9,"); got != "9.9.9.9" {
		t.Errorf("expected 9.9.9.9, got %q", got)
	}
	if got := GetIpFromRequest(server, "1.2.3.4", ",9.9.9.9"); got != "9.9.9.9" {
		t.Errorf("expected 9.9.9.9, got %q", got)
	}
	if got := GetIpFromRequest(server, "1.2.3.4", ",9.9.9.9,"); got != "9.9.9.9" {
		t.Errorf("expected 9.9.9.9, got %q", got)
	}
	if got := GetIpFromRequest(server, "1.2.3.4", ",9.9.9.9,,"); got != "9.9.9.9" {
		t.Errorf("expected 9.9.9.9, got %q", got)
	}

	// x-forwarded-for with trust proxy and "x-forwarded-for" is a private IP
	server.AikidoConfig.TrustProxy = true
	if got := GetIpFromRequest(server, "1.2.3.4", "127.0.0.1"); got != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %q", got)
	}
	if got := GetIpFromRequest(server, "df89:84af:85e0:c55f:960c:341a:2cc6:734d", "::1"); got != "df89:84af:85e0:c55f:960c:341a:2cc6:734d" {
		t.Errorf("expected df89:84af:85e0:c55f:960c:341a:2cc6:734d, got %q", got)
	}

	// x-forwarded-for with trust proxy and "x-forwarded-for" contains private IP
	server.AikidoConfig.TrustProxy = true
	if got := GetIpFromRequest(server, "1.2.3.4", "127.0.0.1, 9.9.9.9"); got != "9.9.9.9" {
		t.Errorf("expected 9.9.9.9, got %q", got)
	}
	if got := GetIpFromRequest(server, "df89:84af:85e0:c55f:960c:341a:2cc6:734d", "::1, a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880"); got != "a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880" {
		t.Errorf("expected a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880, got %q", got)
	}

	// x-forwarded-for with trust proxy and "x-forwarded-for" is public IP
	server.AikidoConfig.TrustProxy = true
	if got := GetIpFromRequest(server, "1.2.3.4", "9.9.9.9"); got != "9.9.9.9" {
		t.Errorf("expected 9.9.9.9, got %q", got)
	}
	if got := GetIpFromRequest(server, "df89:84af:85e0:c55f:960c:341a:2cc6:734d", "a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880"); got != "a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880" {
		t.Errorf("expected a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880, got %q", got)
	}

	// x-forwarded-for with trust proxy and "x-forwarded-for" contains private IP at the end
	server.AikidoConfig.TrustProxy = true
	if got := GetIpFromRequest(server, "1.2.3.4", "9.9.9.9, 127.0.0.1"); got != "9.9.9.9" {
		t.Errorf("expected 9.9.9.9, got %q", got)
	}
	if got := GetIpFromRequest(server, "df89:84af:85e0:c55f:960c:341a:2cc6:734d", "a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880, ::1"); got != "a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880" {
		t.Errorf("expected a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880, got %q", got)
	}

	// x-forwarded-for with trust proxy and multiple IPs
	server.AikidoConfig.TrustProxy = true
	if got := GetIpFromRequest(server, "1.2.3.4", "9.9.9.9, 8.8.8.8, 7.7.7.7"); got != "9.9.9.9" {
		t.Errorf("expected 9.9.9.9, got %q", got)
	}
	if got := GetIpFromRequest(server, "df89:84af:85e0:c55f:960c:341a:2cc6:734d", "a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880, 3b07:2fba:0270:2149:5fc1:2049:5f04:2131, 791d:967e:428a:90b9:8f6f:4fcc:5d88:015d"); got != "a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880" {
		t.Errorf("expected a3ad:8f95:d2a8:454b:cf19:be6e:73c6:f880, got %q", got)
	}

	// x-forwarded-for with trust proxy and many IPs
	server.AikidoConfig.TrustProxy = true
	if got := GetIpFromRequest(server, "1.2.3.4", "127.0.0.1, 192.168.0.1, 192.168.0.2, 9.9.9.9"); got != "9.9.9.9" {
		t.Errorf("expected 9.9.9.9, got %q", got)
	}
	if got := GetIpFromRequest(server, "1.2.3.4", "9.9.9.9, 127.0.0.1, 192.168.0.1, 192.168.0.2"); got != "9.9.9.9" {
		t.Errorf("expected 9.9.9.9, got %q", got)
	}

}
