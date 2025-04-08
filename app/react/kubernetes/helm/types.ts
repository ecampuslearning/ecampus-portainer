export interface Chart extends HelmChartResponse {
  repo: string;
}

export interface HelmChartResponse {
  name: string;
  description: string;
  icon?: string;
  annotations?: {
    category?: string;
  };
}

export interface HelmRepositoryResponse {
  Id: number;
  UserId: number;
  URL: string;
}

export interface HelmRepositoriesResponse {
  GlobalRepository: string;
  UserRepositories: HelmRepositoryResponse[];
}

export interface HelmChartsResponse {
  entries: Record<string, HelmChartResponse[]>;
  apiVersion: string;
  generated: string;
}

export type InstallChartPayload = {
  Name: string;
  Repo: string;
  Chart: string;
  Values: string;
  Namespace: string;
};
