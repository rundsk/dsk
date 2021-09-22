/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import React, { useState, useEffect } from 'react';
import { Client } from '@rundsk/js-sdk';
import './Search.css';
import { withRoute } from 'react-router5';

function SearchResult(props) {
  const ref = React.createRef();

  useEffect(() => {
    if (props.isFocused && ref.current) {
      ref.current.scrollIntoView({ behavior: 'smooth', block: 'end' });
    }
  }, [props.isFocused, ref]);

  function handleClick() {
    props.onSelect();
    props.router.navigate('node', { node: `${props.url}` });
  }

  let snippet = props.description;

  if (props.fragments.length > 0) {
    snippet = props.fragments[0];
  }

  let classes = ['search-result'];

  if (props.isFocused) {
    classes.push('search-result--is-focused');
  }

  return (
    <div ref={ref} className={classes.join(' ')} onClick={handleClick}>
      <div className="search-result__title">{props.title}</div>
      {snippet && <div className="search-result__snippet" dangerouslySetInnerHTML={{ __html: snippet }} />}
      <div className="search-result__path">/{props.url}</div>
    </div>
  );
}

function Search(props) {
  const [searchTerm, setSearchTerm] = useState(props.searchTerm || '');
  const [searchIsFocused, setSearchIsFocused] = useState(false);
  const [shouldShowResults, setShouldShowResults] = useState(false);
  const [searchResults, setSearchResults] = useState([]);
  const [focusedResult, setFocusedResult] = useState(0);

  const searchInputRef = React.createRef();

  const shortcutHandler = (event) => {
    if (event.key === 'ArrowDown' && searchResults.length > 0) {
      event.preventDefault();

      if (focusedResult < searchResults.length - 1) {
        setFocusedResult(focusedResult + 1);
      }
    }

    if (event.key === 'ArrowUp' && searchResults.length > 0) {
      event.preventDefault();

      if (focusedResult > 0) {
        setFocusedResult(focusedResult - 1);
      }
    }

    if (event.key === 'Enter' && searchResults.length > 0) {
      if (searchResults.length > 0 && searchResults.length >= focusedResult - 1) {
        blur();
        setSearchTerm('');
        hideSearch();
        let selectedItem = searchResults[focusedResult];
        props.router.navigate('node', { node: `${selectedItem.url}` });
      }
    }

    if (event.key === 'Escape' && searchIsFocused) {
      event.preventDefault();
      blur();
      hideSearch();
    }

    if (event.key === 's' && event.target.nodeName !== 'INPUT') {
      event.preventDefault();
      focus();
    }
  };

  useEffect(() => {
    document.addEventListener('keydown', shortcutHandler);

    return () => {
      document.removeEventListener('keydown', shortcutHandler);
    };
  });

  useEffect(() => {
    search(searchTerm);
  }, [searchTerm]);

  function onSearchTermChange(ev) {
    setSearchTerm(ev.target.value);
    setFocusedResult(0);
    setShouldShowResults(true);
  }

  async function search(term) {
    if (!term) {
      // No search term given, results in showing the full unfiltered tree (clear).
      setSearchResults([]);
      return;
    }

    const search = Client.search(term);
    search
      .then((data) => {
        if (!data.hits) {
          // Filtering yielded no results, we save us iterating over the
          // existing tree, as we already know what it should look like.
          setSearchResults([]);
          return;
        }
        setSearchResults(data.hits);
      })
      .catch((error) => {
        console.log(error);
      });
  }

  function showSearch() {
    setSearchIsFocused(true);
    setShouldShowResults(searchTerm !== '');
  }

  function hideSearch() {
    setSearchIsFocused(false);
    setShouldShowResults(false);
  }

  function blur() {
    if (searchInputRef.current) {
      searchInputRef.current.blur();
    }
  }

  function focus() {
    if (searchInputRef.current) {
      searchInputRef.current.focus();
    }
  }

  let classes = ['search'];

  if (searchIsFocused) {
    classes.push('search--is-focused');
  }

  return (
    <div
      className={classes.join(' ')}
      onMouseDown={(ev) => {
        if (searchIsFocused) {
          hideSearch();
        }
      }}
    >
      <div className="search__content-container">
        <div
          className="search__content"
          onClick={(ev) => {
            ev.stopPropagation();
          }}
        >
          <input
            type="search"
            placeholder={`Search ${props.title}…`}
            value={searchTerm}
            onChange={onSearchTermChange}
            onFocus={(ev) => {
              ev.preventDefault();
              ev.stopPropagation();
              showSearch();
            }}
            ref={searchInputRef}
            onClick={(ev) => {
              ev.stopPropagation();
              ev.preventDefault();
            }}
          />

          <div
            className={`search__results-container${shouldShowResults ? ' search__results-container--is-visible' : ''}`}
          >
            <div className="search__results">
              {searchResults.map((r, i) => {
                return (
                  <SearchResult
                    {...r}
                    isFocused={focusedResult === i}
                    router={props.router}
                    key={r.url}
                    onSelect={() => {
                      blur();
                      setSearchTerm('');
                      hideSearch();
                    }}
                  />
                );
              })}

              {searchResults.length === 0 && searchTerm !== '' && (
                <div className="search__no-dice">No aspects found :(</div>
              )}

              {searchTerm === '' && <div className="search__no-dice">Start typing to search {props.title}…</div>}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default withRoute(Search);
