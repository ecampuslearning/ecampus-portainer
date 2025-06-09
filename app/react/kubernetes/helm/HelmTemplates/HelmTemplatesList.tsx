import { useState, useMemo } from 'react';
import { components, OptionProps } from 'react-select';

import {
  PortainerSelect,
  Option,
} from '@/react/components/form-components/PortainerSelect';
import { Link } from '@/react/components/Link';

import { InsightsBox } from '@@/InsightsBox';
import { SearchBar } from '@@/datatables/SearchBar';
import { InlineLoader } from '@@/InlineLoader';

import { Chart } from '../types';

import { HelmTemplatesListItem } from './HelmTemplatesListItem';

interface Props {
  isLoading: boolean;
  charts?: Chart[];
  selectAction: (chart: Chart) => void;
  registries: string[];
  selectedRegistry: string | null;
  setSelectedRegistry: (registry: string | null) => void;
}

export function HelmTemplatesList({
  isLoading,
  charts = [],
  selectAction,
  registries,
  selectedRegistry,
  setSelectedRegistry,
}: Props) {
  const [textFilter, setTextFilter] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);

  const categories = useMemo(() => getCategories(charts), [charts]);
  const registryOptions = useMemo(
    () =>
      registries.map((registry) => ({
        label: registry,
        value: registry,
      })),
    [registries]
  );

  const filteredCharts = useMemo(
    () => getFilteredCharts(charts, textFilter, selectedCategory),
    [charts, textFilter, selectedCategory]
  );

  return (
    <section className="datatable" aria-label="Helm charts">
      <div className="toolBar vertical-center relative w-full flex-wrap !gap-x-5 !gap-y-1 !px-0 !overflow-visible">
        <div className="toolBarTitle vertical-center">Helm chart</div>

        <SearchBar
          value={textFilter}
          onChange={(value) => setTextFilter(value)}
          placeholder="Search..."
          data-cy="helm-templates-search"
          className="!mr-0 h-9"
        />

        <div className="w-full sm:w-1/4">
          <PortainerSelect
            placeholder="Select a registry"
            value={selectedRegistry ?? ''}
            options={registryOptions}
            onChange={setSelectedRegistry}
            isClearable
            bindToBody
            components={{ Option: RegistryOption }}
            data-cy="helm-registry-select"
          />
        </div>

        <div className="w-full sm:w-1/4">
          <PortainerSelect
            placeholder="Select a category"
            value={selectedCategory}
            options={categories}
            onChange={setSelectedCategory}
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

        {isLoading && (
          <div className="flex flex-col">
            <InlineLoader className="justify-center">
              Loading helm charts...
            </InlineLoader>
            {charts.length === 0 && (
              <div className="text-muted text-center">
                Initial download of Helm charts can take a few minutes
              </div>
            )}
          </div>
        )}

        {!isLoading && charts.length === 0 && selectedRegistry && (
          <div className="text-muted text-center">
            No helm charts available in this registry.
          </div>
        )}

        {!selectedRegistry && (
          <div className="text-muted text-center">
            Please select a registry to view available Helm charts.
          </div>
        )}
      </div>
    </section>
  );
}

// truncate the registry text, because some registry names are urls, which are too long
function RegistryOption(props: OptionProps<Option<string>>) {
  const { data: registry } = props;

  return (
    <div title={registry.value}>
      {/* eslint-disable-next-line react/jsx-props-no-spreading */}
      <components.Option {...props} className="whitespace-nowrap truncate">
        {registry.value}
      </components.Option>
    </div>
  );
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
