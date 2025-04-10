import Markdown from 'markdown-to-jsx';

type Props = {
  notes: string;
};

export function NotesDetails({ notes }: Props) {
  return <Markdown className="list-inside mt-6">{notes}</Markdown>;
}
