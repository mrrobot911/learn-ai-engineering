// @vitest-environment jsdom
import { afterEach, describe, expect, it } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { Markdown } from './Markdown';

afterEach(() => {
  cleanup();
});

describe('Markdown', () => {
  it('renders a body that starts with # followed by a paragraph, ## heading, and list', () => {
    const source =
      '# Intro\n\nThis is a paragraph.\n\n## Subheading\n\n- item 1\n- item 2';
    render(<Markdown source={source} />);

    const headings = screen.getAllByRole('heading', { level: 2 });
    expect(headings.map((h) => h.textContent)).toEqual(['Intro', 'Subheading']);
    expect(screen.getByText('This is a paragraph.')).toBeTruthy();
    expect(screen.getByText('item 1')).toBeTruthy();
    expect(screen.getByText('item 2')).toBeTruthy();
  });

  it('renders a body that is only a single # title', () => {
    render(<Markdown source="# Only a title" />);

    const heading = screen.getByRole('heading', { level: 2 });
    expect(heading.textContent).toBe('Only a title');
  });
});
