import clsx from 'clsx';
import { PropsWithChildren } from 'react';
import _ from 'lodash';
import { Info } from 'lucide-react';

import { truncate } from '@/portainer/filters/filters';
import { UserId } from '@/portainer/users/types';
import { TeamId } from '@/react/portainer/users/teams/types';
import { useTeams } from '@/react/portainer/users/teams/queries';
import { useUsers } from '@/portainer/users/queries';
import { pluralize } from '@/portainer/helpers/strings';
import { ownershipIcon } from '@/react/docker/components/datatable/createOwnershipColumn';

import { Link } from '@@/Link';
import { Tooltip } from '@@/Tip/Tooltip';
import { Icon } from '@@/Icon';

import {
  ResourceControlOwnership,
  ResourceControlType,
  ResourceId,
} from '../types';
import { ResourceControlViewModel } from '../models/ResourceControlViewModel';

interface Props {
  resourceControl?: ResourceControlViewModel;
  resourceType: ResourceControlType;
  isAuthorisedToFetchUsers?: boolean;
}

export function AccessControlPanelDetails({
  resourceControl,
  resourceType,
  isAuthorisedToFetchUsers = false,
}: Props) {
  const inheritanceMessage = getInheritanceMessage(
    resourceType,
    resourceControl
  );

  const {
    Ownership: ownership = ResourceControlOwnership.ADMINISTRATORS,
    UserAccesses: restrictedToUsers = [],
    TeamAccesses: restrictedToTeams = [],
  } = resourceControl || {};

  const users = useAuthorizedUsers(
    restrictedToUsers.map((ra) => ra.UserId),
    isAuthorisedToFetchUsers
  );
  const teams = useAuthorizedTeams(restrictedToTeams.map((ra) => ra.TeamId));

  const teamsLength = teams.data ? teams.data.length : 0;
  const unauthoisedTeams = restrictedToTeams.length - teamsLength;

  let teamsMessage = teams.data && teams.data.join(', ');
  if (unauthoisedTeams > 0 && teams.isFetched) {
    teamsMessage += teamsLength > 0 ? ' and' : '';
    teamsMessage += ` ${unauthoisedTeams} ${pluralize(
      unauthoisedTeams,
      'team'
    )} you are not part of`;
  }

  const userMessage = users.data
    ? users.data.join(', ')
    : `${restrictedToUsers.length} ${pluralize(
        restrictedToUsers.length,
        'user'
      )}`;

  return (
    <table className="table">
      <tbody>
        <tr data-cy="access-ownership">
          <td className="w-1/5">Ownership</td>
          <td>
            <i
              className={clsx(ownershipIcon(ownership), 'space-right')}
              aria-hidden="true"
              aria-label="ownership-icon"
            />
            <span aria-label="ownership">{ownership}</span>
            <Tooltip message={getOwnershipTooltip(ownership)} />
          </td>
        </tr>
        {inheritanceMessage}
        {restrictedToUsers.length > 0 && (
          <tr data-cy="access-authorisedUsers">
            <td>Authorized users</td>
            <td aria-label="authorized-users">{userMessage}</td>
          </tr>
        )}
        {restrictedToTeams.length > 0 && (
          <tr data-cy="access-authorisedTeams">
            <td>Authorized teams</td>
            <td aria-label="authorized-teams">{teamsMessage}</td>
          </tr>
        )}
      </tbody>
    </table>
  );
}

function getOwnershipTooltip(ownership: ResourceControlOwnership) {
  switch (ownership) {
    case ResourceControlOwnership.PRIVATE:
      return 'Management of this resource is restricted to a single user.';
    case ResourceControlOwnership.RESTRICTED:
      return 'This resource can be managed by a restricted set of users and/or teams.';
    case ResourceControlOwnership.PUBLIC:
      return 'This resource can be managed by any user with access to this environment.';
    case ResourceControlOwnership.ADMINISTRATORS:
    default:
      return 'This resource can only be managed by administrators.';
  }
}

function getInheritanceMessage(
  resourceType: ResourceControlType,
  resourceControl?: ResourceControlViewModel
) {
  if (!resourceControl || resourceControl.Type === resourceType) {
    return null;
  }

  const parentType = resourceControl.Type;
  const resourceId = resourceControl.ResourceId;

  if (
    resourceType === ResourceControlType.Container &&
    parentType === ResourceControlType.Service
  ) {
    return (
      <InheritanceMessage tooltip="Access control applied on a service is also applied on each container of that service.">
        Access control on this resource is inherited from the following service:
        <Link
          to="docker.services.service"
          params={{ id: resourceId }}
          data-cy="docker-access-inherited-service"
        >
          {truncate(resourceId)}
        </Link>
      </InheritanceMessage>
    );
  }

  if (
    resourceType === ResourceControlType.Volume &&
    parentType === ResourceControlType.Container
  ) {
    return (
      <InheritanceMessage tooltip="Access control applied on a container created using a template is also applied on each volume associated to the container.">
        Access control on this resource is inherited from the following
        container:
        <Link
          to="docker.containers.container"
          params={{ id: resourceId }}
          data-cy="docker-access-inherited-container"
        >
          {truncate(resourceId)}
        </Link>
      </InheritanceMessage>
    );
  }

  if (parentType === ResourceControlType.Stack) {
    return (
      <InheritanceMessage tooltip="Access control applied on a stack is also applied on each resource in the stack.">
        <span className="space-right">
          Access control on this resource is inherited from the following stack:
        </span>
        {removeEndpointIdFromStackResourceId(resourceId)}
      </InheritanceMessage>
    );
  }

  return null;
}

function removeEndpointIdFromStackResourceId(stackName: ResourceId) {
  if (!stackName || typeof stackName !== 'string') {
    return stackName;
  }

  const firstUnderlineIndex = stackName.indexOf('_');
  if (firstUnderlineIndex < 0) {
    return stackName;
  }
  return stackName.substring(firstUnderlineIndex + 1);
}

interface InheritanceMessageProps {
  tooltip: string;
}

function InheritanceMessage({
  children,
  tooltip,
}: PropsWithChildren<InheritanceMessageProps>) {
  return (
    <tr>
      <td colSpan={2} aria-label="inheritance-message">
        <div className="inline-flex items-center gap-1">
          <Icon icon={Info} mode="primary" />
          {children}
        </div>
        <Tooltip message={tooltip} />
      </td>
    </tr>
  );
}

function useAuthorizedTeams(authorizedTeamIds: TeamId[]) {
  return useTeams(false, 0, {
    enabled: authorizedTeamIds.length > 0,
    select: (teams) => {
      if (authorizedTeamIds.length === 0) {
        return [];
      }

      return _.compact(
        authorizedTeamIds.map((id) => {
          const team = teams.find((u) => u.Id === id);
          return team?.Name;
        })
      );
    },
  });
}

function useAuthorizedUsers(authorizedUserIds: UserId[], enabled = true) {
  return useUsers(
    false,
    0,
    authorizedUserIds.length > 0 && enabled,
    (users) => {
      if (authorizedUserIds.length === 0) {
        return [];
      }

      return _.compact(
        authorizedUserIds.map((id) => {
          const user = users.find((u) => u.Id === id);
          return user?.Username;
        })
      );
    }
  );
}
