/**
 * Copyright 2023 Gravitational, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { assertUnreachable } from 'teleterm/ui/utils';

import {
  LabelMatch,
  ResourceMatch,
  SearchResult,
  mainResourceName,
  mainResourceField,
  searchableFields,
} from './searchResult';

/**
 * useSearch returns a function which searches for the given list of space-separated keywords across
 * all root and leaf clusters that the user is currently logged in to.
 *
 * It does so by issuing a separate request for each resource type to each cluster. It fails if any
 * of those requests fail.
 */
export async function searchResources(
  clustersService,
  resourcesService,
  search: string
): Promise<SearchResult[]> {
  const connectedClusters = clustersService
    .getClusters()
    .filter(c => c.connected);
  const searchPromises = connectedClusters.map(cluster =>
    resourcesService.searchResources(cluster.uri, search)
  );
  const searchResults = (await Promise.all(searchPromises)).flat();

  return sortResults(searchResults, search).slice(0, 10);
}

export function sortResults(
  searchResults: SearchResult[],
  search: string
): SearchResult[] {
  const terms = search
    .split(' ')
    .filter(Boolean)
    // We have to match the implementation of the search algorithm as closely as possible. It uses
    // strings.ToLower from Go which unfortunately doesn't have a good equivalent in JavaScript.
    //
    // strings.ToLower uses some kind of a universal map for lowercasing non-ASCII characters such
    // as the Turkish İ. JavaScript doesn't have such a function, possibly because it's not possible
    // to have universal case mapping. [1]
    //
    // The closest thing that JS has is toLocaleLowerCase. Since we don't know what locale the
    // search string uses, we let the runtime figure it out based on the system settings.
    // The assumption is that if someone has a resource with e.g. Turkish characters, their system
    // is set to the appropriate locale and the search results will be properly scored.
    //
    // Highlighting will have problems with some non-ASCII characters anyway because the library we
    // use for highlighting uses a regex with the i flag underneath.
    //
    // [1] https://web.archive.org/web/20190113111936/https://blogs.msdn.microsoft.com/oldnewthing/20030905-00/?p=42643
    .map(term => term.toLocaleLowerCase());
  const collator = new Intl.Collator();

  return searchResults
    .map(searchResult => calculateScore(populateMatches(searchResult, terms)))
    .sort(
      (a, b) =>
        // Highest score first.
        b.score - a.score ||
        collator.compare(mainResourceName(a), mainResourceName(b))
    );
}

function populateMatches(
  searchResult: SearchResult,
  terms: string[]
): SearchResult {
  const labelMatches: LabelMatch[] = [];
  const resourceMatches: ResourceMatch<SearchResult['kind']>[] = [];

  terms.forEach(term => {
    searchResult.resource.labelsList.forEach(label => {
      // indexOf is faster on Chrome than includes or regex.
      // https://jsbench.me/b7lf9kvrux/1
      const nameIndex = label.name.toLocaleLowerCase().indexOf(term);
      const valueIndex = label.value.toLocaleLowerCase().indexOf(term);

      if (nameIndex >= 0) {
        labelMatches.push({
          kind: 'label-name',
          labelName: label.name,
          searchTerm: term,
        });
      }

      if (valueIndex >= 0) {
        labelMatches.push({
          kind: 'label-value',
          labelName: label.name,
          searchTerm: term,
        });
      }
    });

    searchableFields[searchResult.kind].forEach(field => {
      // `String` here is just to satisfy the compiler.
      const index = searchResult.resource[field]
        .toLocaleLowerCase()
        .indexOf(term);

      if (index >= 0) {
        resourceMatches.push({
          field,
          searchTerm: term,
        });
      }
    });
  });

  return { ...searchResult, labelMatches, resourceMatches };
}

// TODO(ravicious): Extract the scoring logic to a function to better illustrate different weight
// for different matches.
function calculateScore(searchResult: SearchResult): SearchResult {
  let totalScore = 0;

  for (const match of searchResult.labelMatches) {
    const { searchTerm } = match;
    switch (match.kind) {
      case 'label-name': {
        const label = searchResult.resource.labelsList.find(
          label => label.name === match.labelName
        );
        const score = Math.floor((searchTerm.length / label.name.length) * 100);
        totalScore += score;
        break;
      }
      case 'label-value': {
        const label = searchResult.resource.labelsList.find(
          label => label.name === match.labelName
        );
        const score = Math.floor(
          (searchTerm.length / label.value.length) * 100
        );
        totalScore += score;
        break;
      }
      default: {
        assertUnreachable(match.kind);
      }
    }
  }

  for (const match of searchResult.resourceMatches) {
    const { searchTerm } = match;
    const field = searchResult.resource[match.field];
    const isMainField = mainResourceField[searchResult.kind] === match.field;
    const weight = isMainField ? 4 : 2;

    const score = Math.floor((searchTerm.length / field.length) * 100 * weight);
    totalScore += score;
  }

  return { ...searchResult, score: totalScore };
}
