export function mockCodeMirror() {
  vi.mock('@uiw/react-codemirror', () => ({
    __esModule: true,
    default: () => <div />,
    oneDarkHighlightStyle: {},
    keymap: {
      of: () => ({}),
    },
  }));
  vi.mock('yaml-schema', () => ({
    yamlSchema: () => [],
    validation: () => ({
      of: () => ({}),
    }),
  }));
}
