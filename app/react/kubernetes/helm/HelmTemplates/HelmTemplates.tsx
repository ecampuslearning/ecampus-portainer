import { useState } from 'react';

import { useCurrentUser } from '@/react/hooks/useUser';

import { Chart } from '../types';

import { useHelmChartList } from './queries/useHelmChartList';
import { HelmTemplatesList } from './HelmTemplatesList';
import { HelmTemplatesSelectedItem } from './HelmTemplatesSelectedItem';

interface Props {
  onSelectHelmChart: (chartName: string) => void;
  namespace?: string;
  name?: string;
}

export function HelmTemplates({ onSelectHelmChart, namespace, name }: Props) {
  const [selectedChart, setSelectedChart] = useState<Chart | null>(null);

  const { user } = useCurrentUser();
  const { data: charts = [], isLoading: chartsLoading } = useHelmChartList(
    user.Id
  );

  function clearHelmChart() {
    setSelectedChart(null);
    onSelectHelmChart('');
  }

  function handleChartSelection(chart: Chart) {
    setSelectedChart(chart);
    onSelectHelmChart(chart.name);
  }

  return (
    <div className="row">
      <div className="col-sm-12 p-0">
        {selectedChart ? (
          <HelmTemplatesSelectedItem
            selectedChart={selectedChart}
            clearHelmChart={clearHelmChart}
            namespace={namespace}
            name={name}
          />
        ) : (
          <HelmTemplatesList
            charts={charts}
            selectAction={handleChartSelection}
            loading={chartsLoading}
          />
        )}
      </div>
    </div>
  );
}
