import { Checkbox } from '@@/form-components/Checkbox';
import { CodeEditor } from '@@/CodeEditor';

import { Values } from '../../types';

interface Props {
  values?: Values;
  isUserSupplied: boolean;
  setIsUserSupplied: (isUserSupplied: boolean) => void;
}

const noValuesMessage = 'No values found';

export function ValuesDetails({
  values,
  isUserSupplied,
  setIsUserSupplied,
}: Props) {
  return (
    <div className="relative">
      {/* bring in line with the code editor copy button */}
      <div className="absolute top-1 left-0">
        <Checkbox
          label="User defined only"
          id="values-details-user-supplied"
          checked={isUserSupplied}
          onChange={() => setIsUserSupplied(!isUserSupplied)}
          data-cy="values-details-user-supplied"
        />
      </div>
      <CodeEditor
        type="yaml"
        id="values-details-code-editor"
        data-cy="values-details-code-editor"
        value={
          isUserSupplied
            ? values?.userSuppliedValues ?? noValuesMessage
            : values?.computedValues ?? noValuesMessage
        }
        readonly
      />
    </div>
  );
}
