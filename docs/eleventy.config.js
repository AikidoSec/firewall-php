import syntaxHighlight from "@11ty/eleventy-plugin-syntaxhighlight";

export default async function(eleventyConfig) {
  const getNavPages = (collectionApi) => {
    return collectionApi
      .getAll()
      .filter((item) => {
        // Keep only real pages with output, and skip common utility pages
        if (!item.url || item.url.startsWith("/assets/")) return false;
        if (item.data && item.data.eleventyExcludeFromCollections) return false;
        if (item.data && item.data.favorite === true) return false; // only show favorites
        if (item.url === "/404/") return false;
        return true;
      })
      .map((item) => {
        const inputPath = (item.inputPath || "")
          .replace(/^\.\//, "")
          .replace(/\\/g, "/");
        const parts = inputPath.split("/");
        const navDir = parts.length > 1 ? parts.slice(0, -1).join("/") : "";
        item.data = item.data || {};
        item.data.navDir = navDir;
        return item;
      })
      .sort((a, b) => {
        const aTitle = (a.data.title || a.fileSlug || "").toLowerCase();
        const bTitle = (b.data.title || b.fileSlug || "").toLowerCase();
        return aTitle.localeCompare(bTitle, "en");
      });
  };

  eleventyConfig.addPlugin(syntaxHighlight);
  eleventyConfig.addPassthroughCopy({ "_includes/styles.css": "styles.css" });
  eleventyConfig.addGlobalData("layout", "default.njk");
  eleventyConfig.addGlobalData("agent", "PHP");
  eleventyConfig.addCollection("pages", (collectionApi) => getNavPages(collectionApi));
  eleventyConfig.addCollection("nav", (collectionApi) => {
    const pages = getNavPages(collectionApi);
    const groups = new Map();

    for (const page of pages) {
      const dir = page.data.navDir || "";
      if (!groups.has(dir)) groups.set(dir, []);
      groups.get(dir).push(page);
    }

    const entries = [...groups.entries()].sort(([a], [b]) => {
      if (!a && b) return 1;
      if (a && !b) return -1;
      return a.localeCompare(b, "en");
    });

    return entries.map(([dir, list]) => ({ dir, pages: list }));
  });
};
