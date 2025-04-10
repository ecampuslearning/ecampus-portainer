import { CodeEditor } from '@@/CodeEditor';

type Props = {
  manifest: string;
};

export function ManifestDetails({ manifest }: Props) {
  return (
    <CodeEditor
      id="helm-manifest"
      type="yaml"
      data-cy="helm-manifest"
      value={manifest}
      height="600px"
      readonly
    />
  );
}
