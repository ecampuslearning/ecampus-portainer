export function mockCodeMirror() {
  vi.mock('@uiw/react-codemirror', () => ({
    __esModule: true,
    default: () => <div />,
    oneDarkHighlightStyle: {},
    keymap: {
      of: () => ({}),
    },
  }));

  vi.mock('react-codemirror-merge', () => {
    const components = {
      MergeView: () => <div />,
      Original: () => <div />,
      Modified: () => <div />,
    };

    return {
      __esModule: true,
      default: components,
      ...components,
    };
  });

  vi.mock('yaml-schema', () => ({
    yamlSchema: () => [],
    validation: () => ({
      of: () => ({}),
    }),
  }));
}
