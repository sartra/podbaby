import React from 'react';
import { Provider } from 'react-redux';
import DocumentTitle from 'react-document-title';
import { getTitle } from './utils';
import DevTools from './devtools';


export default class Root extends React.Component {

  render() {

    const { store, routes } = this.props;

    return (
      <DocumentTitle title={getTitle()}>
        <div>
        <Provider store={store}>
          {routes}
        </Provider>
        <DevTools store={store} />
        </div>
      </DocumentTitle>
    );
  }

}

