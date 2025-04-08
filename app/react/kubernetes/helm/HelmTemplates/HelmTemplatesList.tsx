import { useState, useMemo } from 'react';

import { PortainerSelect } from '@/react/components/form-components/PortainerSelect';
import { Link } from '@/react/components/Link';

import { InsightsBox } from '@@/InsightsBox';
import { SearchBar } from '@@/datatables/SearchBar';

import { Chart } from '../types';

import { HelmTemplatesListItem } from './HelmTemplatesListItem';

interface Props {
  loading: boolean;
  charts?: Chart[];
  selectAction: (chart: Chart) => void;
}

/**
 * Get categories from charts
 * @param charts - The charts to get the categories from
 * @returns Categories
 */
function getCategories(charts: Chart[]) {
  const annotationCategories = charts
    .map((chart) => chart.annotations?.category) // get category
    .filter((c): c is string => !!c); // filter out nulls/undefined

  const availableCategories = [...new Set(annotationCategories)].sort(); // unique and sort

  // Create options array in the format expected by PortainerSelect
  return availableCategories.map((cat) => ({
    label: cat,
    value: cat,
  }));
}

/**
 * Get filtered charts
 * @param charts - The charts to get the filtered charts from
 * @param textFilter - The text filter
 * @param selectedCategory - The selected category
 * @returns Filtered charts
 */
function getFilteredCharts(
  charts: Chart[],
  textFilter: string,
  selectedCategory: string | null
) {
  return charts.filter((chart) => {
    // Text filter
    if (
      textFilter &&
      !chart.name.toLowerCase().includes(textFilter.toLowerCase()) &&
      !chart.description.toLowerCase().includes(textFilter.toLowerCase())
    ) {
      return false;
    }

    // Category filter
    if (
      selectedCategory &&
      (!chart.annotations || chart.annotations.category !== selectedCategory)
    ) {
      return false;
    }

    return true;
  });
}

export function HelmTemplatesList({
  loading,
  charts = [],
  selectAction,
}: Props) {
  const [textFilter, setTextFilter] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);

  const categories = useMemo(() => getCategories(charts), [charts]);

  const filteredCharts = useMemo(
    () => getFilteredCharts(charts, textFilter, selectedCategory),
    [charts, textFilter, selectedCategory]
  );

  return (
    <section className="datatable" aria-label="Helm charts">
      <div className="toolBar vertical-center relative w-full flex-wrap !gap-x-5 !gap-y-1 !px-0">
        <div className="toolBarTitle vertical-center">Helm chart</div>

        <SearchBar
          value={textFilter}
          onChange={(value) => setTextFilter(value)}
          placeholder="Search..."
          data-cy="helm-templates-search"
          className="!mr-0 h-9"
        />

        <div className="w-full sm:w-1/5">
          <PortainerSelect
            placeholder="Select a category"
            value={selectedCategory}
            options={categories}
            onChange={(value) => setSelectedCategory(value)}
            isClearable
            bindToBody
            data-cy="helm-category-select"
          />
        </div>
      </div>
      <div className="w-fit">
        <div className="small text-muted mb-2">
          Select the Helm chart to use. Bring further Helm charts into your
          selection list via{' '}
          <Link
            to="portainer.account"
            params={{ '#': 'helm-repositories' }}
            data-cy="helm-repositories-link"
          >
            User settings - Helm repositories
          </Link>
          .
        </div>

        <InsightsBox
          header="Disclaimer"
          type="slim"
          content={
            <>
              At present Portainer does not support OCI format Helm charts.
              Support for OCI charts will be available in a future release.
              <br />
              If you would like to provide feedback on OCI support or get access
              to early releases to test this functionality,{' '}
              <a
                href="https://bit.ly/3WVkayl"
                target="_blank"
                rel="noopener noreferrer"
              >
                please get in touch
              </a>
              .
            </>
          }
        />
      </div>

      <div className="blocklist !px-0" role="list">
        {filteredCharts.map((chart) => (
          <HelmTemplatesListItem
            key={chart.name}
            model={chart}
            onSelect={selectAction}
          />
        ))}

        {filteredCharts.length === 0 && textFilter && (
          <div className="text-muted small mt-4">No Helm charts found</div>
        )}

        {loading && (
          <div className="text-muted text-center">
            Loading...
            <div className="text-muted text-center">
              Initial download of Helm charts can take a few minutes
            </div>
          </div>
        )}

        {!loading && charts.length === 0 && (
          <div className="text-muted text-center">
            No helm charts available.
          </div>
        )}
      </div>
    </section>
  );
}
