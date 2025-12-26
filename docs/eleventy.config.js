import syntaxHighlight from "@11ty/eleventy-plugin-syntaxhighlight";
import eleventyNavigationPlugin from "@11ty/eleventy-navigation";
import { RenderPlugin } from "@11ty/eleventy";
import { eleventyImageTransformPlugin } from "@11ty/eleventy-img";
import MarkdownItGitHubAlerts from 'markdown-it-github-alerts'

export default async function(eleventyConfig) {
  const markNavState = (items, currentUrl) => {
    if (!Array.isArray(items)) {
      return [];
    }

    return items.map((item) => {
      const children = markNavState(item.children || [], currentUrl);
      const isCurrent = item.url === currentUrl;
      const isOpen = isCurrent || children.some((child) => child.isCurrent || child.isOpen);

      return { ...item, children, isCurrent, isOpen };
    });
  };

  const findNavItem = (items, currentUrl) => {
    if (!Array.isArray(items)) {
      return null;
    }

    for (const item of items) {
      if (item.url === currentUrl) {
        return item;
      }

      const childMatch = findNavItem(item.children || [], currentUrl);
      if (childMatch) {
        return childMatch;
      }
    }

    return null;
  };

  const findBreadcrumb = (items, currentUrl) => {
    if (!Array.isArray(items)) {
      return [];
    }

    for (const item of items) {
      if (item.url === currentUrl) {
        return [item];
      }

      const childTrail = findBreadcrumb(item.children || [], currentUrl);
      if (childTrail.length) {
        return [item, ...childTrail];
      }
    }

    return [];
  };

  // Site specific metadata
  eleventyConfig.addGlobalData("agent", "PHP");

  // Plugins
  eleventyConfig.addPlugin(syntaxHighlight);
  eleventyConfig.addPlugin(RenderPlugin);
  eleventyConfig.addPlugin(eleventyNavigationPlugin);
  eleventyConfig.addPlugin(eleventyImageTransformPlugin);

  // Additional markdown rendering
  eleventyConfig.amendLibrary("md", (mdLib) => mdLib.use(MarkdownItGitHubAlerts));

  // Filters
  eleventyConfig.addFilter("navWithActive", (items, currentUrl) => markNavState(items, currentUrl));
  eleventyConfig.addFilter("navBreadcrumb", (items, currentUrl) => findBreadcrumb(items, currentUrl));
  eleventyConfig.addFilter("navFind", (items, currentUrl) => findNavItem(items, currentUrl));

  // Layout
  eleventyConfig.addPassthroughCopy({ "_includes/styles.css": "styles.css" });
  eleventyConfig.addGlobalData("layout", "default.njk");
};
