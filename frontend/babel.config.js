const isTest = process.env.NODE_ENV !== 'test';

module.exports = {
  presets: ['@babel/preset-env', 'react-app'],
  env: {
    testing: {
      presets: [['@babel/preset-env', !isTest ? {} : { targets: { node: 'current' } }]],
    },
  },
};
