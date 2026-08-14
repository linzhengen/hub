import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

/**
 * Renders Markdown written by the assistant.
 *
 * **Raw HTML is never rendered.** `react-markdown` does not parse HTML into
 * nodes unless `rehype-raw` is added, so there is no sanitiser here to get
 * wrong or to fall behind - the unsafe thing is simply not built. That matters
 * more than usual for this text: the assistant answers partly from tool
 * results, and those carry group descriptions, resource metadata and user names
 * that other people wrote. **Do not add `rehype-raw`.**
 *
 * Links are likewise left to `react-markdown`'s default URL handling, which
 * drops `javascript:` and other executable schemes, and they open in a new tab
 * so a click never navigates away from an unsaved conversation.
 *
 * Styling is done through `components` rather than a typography plugin: the
 * chat is the only Markdown on the site, and one prose stylesheet for one
 * component is not worth another dependency.
 */
export const Markdown: React.FC<{ children: string }> = ({ children }) => (
  <div className="text-sm leading-relaxed break-words">
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
        h1: ({ children }) => <h2 className="mt-3 mb-2 text-base font-bold first:mt-0">{children}</h2>,
        h2: ({ children }) => <h3 className="mt-3 mb-2 text-base font-bold first:mt-0">{children}</h3>,
        h3: ({ children }) => <h4 className="mt-3 mb-1 text-sm font-bold first:mt-0">{children}</h4>,
        ul: ({ children }) => <ul className="mb-2 ml-5 list-disc last:mb-0">{children}</ul>,
        ol: ({ children }) => <ol className="mb-2 ml-5 list-decimal last:mb-0">{children}</ol>,
        li: ({ children }) => <li className="mb-0.5">{children}</li>,
        a: ({ href, children }) => (
          <a href={href} target="_blank" rel="noopener noreferrer" className="underline">
            {children}
          </a>
        ),
        blockquote: ({ children }) => (
          <blockquote className="mb-2 border-l-2 border-gray-300 pl-3 opacity-80 last:mb-0 dark:border-gray-600">
            {children}
          </blockquote>
        ),
        // A fenced block is the one thing here that can be wider than the
        // bubble - a permission id, a JSON result - so it scrolls on its own
        // rather than stretching the transcript.
        pre: ({ children }) => (
          <pre className="mb-2 overflow-x-auto rounded bg-black/5 p-2 text-xs last:mb-0 dark:bg-white/10">
            {children}
          </pre>
        ),
        code: ({ className, children }) => {
          // react-markdown gives a fenced block a language class and an inline
          // span none, which is the only way to tell them apart here.
          const fenced = className !== undefined;
          if (fenced) return <code className="font-mono">{children}</code>;
          return <code className="rounded bg-black/10 px-1 py-0.5 font-mono text-xs dark:bg-white/15">{children}</code>;
        },
        table: ({ children }) => (
          <div className="mb-2 overflow-x-auto last:mb-0">
            <table className="w-full border-collapse text-xs">{children}</table>
          </div>
        ),
        th: ({ children }) => (
          <th className="border border-gray-300 px-2 py-1 text-left font-semibold dark:border-gray-600">
            {children}
          </th>
        ),
        td: ({ children }) => (
          <td className="border border-gray-300 px-2 py-1 dark:border-gray-600">{children}</td>
        ),
        hr: () => <hr className="my-3 border-gray-300 dark:border-gray-600" />,
      }}
    >
      {children}
    </ReactMarkdown>
  </div>
);
