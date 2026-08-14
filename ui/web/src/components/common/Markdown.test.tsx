import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { Markdown } from '@/components/common/Markdown';

describe('Markdown', () => {
  it('renders the structure the assistant lays an answer out with', () => {
    const { container } = render(
      <Markdown>{'## Groups\n\n- admin\n- staff\n\n`role_id`'}</Markdown>,
    );

    expect(screen.getByText('Groups')).toBeTruthy();
    expect(container.querySelectorAll('li')).toHaveLength(2);
    expect(container.querySelector('code')?.textContent).toBe('role_id');
  });

  it('renders GFM tables, which is how it reports a list of things', () => {
    const { container } = render(
      <Markdown>{'| name | id |\n| --- | --- |\n| admin | g1 |'}</Markdown>,
    );

    expect(container.querySelector('table')).toBeTruthy();
    expect(screen.getByText('admin')).toBeTruthy();
  });

  it('renders a fenced block as code', () => {
    const { container } = render(<Markdown>{'```json\n{"id":"g1"}\n```'}</Markdown>);

    const pre = container.querySelector('pre');
    expect(pre).toBeTruthy();
    expect(pre?.textContent).toContain('{"id":"g1"}');
  });

  // The assistant answers partly from tool results, and those carry text other
  // people wrote - group descriptions, resource metadata, user names. None of
  // it may become markup.
  it('never renders raw HTML', () => {
    const { container } = render(
      <Markdown>{'Found this description: <img src=x onerror="alert(1)"> and <b>bold</b>'}</Markdown>,
    );

    expect(container.querySelector('img')).toBeNull();
    expect(container.querySelector('b')).toBeNull();
    // It is shown as the text it is, so the reader can see what was written.
    expect(container.textContent).toContain('onerror');
  });

  it('does not render a script tag', () => {
    const { container } = render(<Markdown>{'<script>alert(1)</script>'}</Markdown>);

    expect(container.querySelector('script')).toBeNull();
  });

  // A link's text is the model's, and its href may have come from a tool
  // result. An executable scheme must not survive.
  it('drops a javascript: link', () => {
    const { container } = render(<Markdown>{'[click me](javascript:alert(1))'}</Markdown>);

    const link = container.querySelector('a');
    expect(link?.getAttribute('href') ?? '').not.toContain('javascript:');
  });

  it('opens an ordinary link in a new tab without leaking the opener', () => {
    const { container } = render(<Markdown>{'[docs](https://example.com/docs)'}</Markdown>);

    const link = container.querySelector('a');
    expect(link?.getAttribute('href')).toBe('https://example.com/docs');
    expect(link?.getAttribute('target')).toBe('_blank');
    expect(link?.getAttribute('rel')).toContain('noopener');
  });

  // A streamed answer arrives a piece at a time, so it is rendered mid-syntax
  // on every frame. That must not throw.
  it('renders a half-finished fenced block without failing', () => {
    const { container } = render(<Markdown>{'Here you go:\n\n```json\n{"id":'}</Markdown>);

    expect(container.textContent).toContain('Here you go:');
  });
});
