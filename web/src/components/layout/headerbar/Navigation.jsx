/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import SkeletonWrapper from '../components/SkeletonWrapper';

const Navigation = ({
  mainNavLinks,
  isMobile,
  isLoading,
  userState,
  pricingRequireAuth,
}) => {
  const location = useLocation();

  const isActive = (path) => {
    if (path === '/') {
      return location.pathname === '/';
    }
    return location.pathname.startsWith(path);
  };

  const renderNavLinks = () => {
    return mainNavLinks.map((link) => {
      let targetPath = link.to;
      if (link.itemKey === 'console' && !userState.user) {
        targetPath = '/login';
      }
      if (link.itemKey === 'pricing' && pricingRequireAuth && !userState.user) {
        targetPath = '/login';
      }

      const active = !link.isExternal && isActive(targetPath);

      const linkClasses = [
        'relative flex-shrink-0 flex items-center justify-center',
        'text-sm font-medium tracking-wider',
        'transition-all duration-300 ease-out',
        active ? 'nav-link-active' : 'nav-link-default',
      ].join(' ');

      const content = <span>{link.text}</span>;

      if (link.isExternal) {
        return (
          <a
            key={link.itemKey}
            href={link.externalLink}
            target='_blank'
            rel='noopener noreferrer'
            className={linkClasses}
          >
            {content}
          </a>
        );
      }

      return (
        <Link key={link.itemKey} to={targetPath} className={linkClasses}>
          {content}
        </Link>
      );
    });
  };

  return (
    <nav className='flex flex-1 items-center justify-center mx-2 md:mx-4 overflow-x-auto whitespace-nowrap scrollbar-hide'>
      <SkeletonWrapper
        loading={isLoading}
        type='navigation'
        count={4}
        width={60}
        height={16}
        isMobile={isMobile}
      >
        <div className='flex items-center gap-x-10'>
          {renderNavLinks()}
        </div>
      </SkeletonWrapper>
    </nav>
  );
};

export default Navigation;
