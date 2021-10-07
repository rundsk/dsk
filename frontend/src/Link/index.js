/**
 * Copyright 2020 Marius Wilms, Christoph Labacher. All rights reserved.
 * Copyright 2019 Atelier Disko. All rights reserved.
 *
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

import { NavLink as RouterNavLink, Link as RouterLink } from 'react-router-dom';

import { constructURL } from '../utils';

export const NavLink = ({ children, to, ...props }) => {
  return (
    <RouterNavLink to={constructURL({ node: to, activeTab: null })} {...props}>
      {children}
    </RouterNavLink>
  );
};

export const Link = ({ children, to, ...props }) => {
  return (
    <RouterLink to={constructURL({ node: to, activeTab: null })} {...props}>
      {children}
    </RouterLink>
  );
};
