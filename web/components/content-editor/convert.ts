import MarkdownIt from "markdown-it"
import TurndownService from "turndown"

const markdownIt = new MarkdownIt({
  html: true,
  linkify: true,
  breaks: true,
})

const turndownService = new TurndownService({
  headingStyle: "atx",
  codeBlockStyle: "fenced",
  bulletListMarker: "-",
  emDelimiter: "*",
  strongDelimiter: "**",
})

turndownService.keep(["table", "thead", "tbody", "tr", "th", "td"])

export function markdownToHtml(markdown: string) {
  return markdownIt.render(markdown ?? "")
}

export function htmlToMarkdown(html: string) {
  return turndownService.turndown(html ?? "")
}
