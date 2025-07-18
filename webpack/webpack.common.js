const path = require('path');
const { ProvidePlugin, IgnorePlugin } = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const WebpackBuildNotifierPlugin = require('webpack-build-notifier');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const { CleanWebpackPlugin } = require('clean-webpack-plugin');
const LodashModuleReplacementPlugin = require('lodash-webpack-plugin');
const TsconfigPathsPlugin = require('tsconfig-paths-webpack-plugin');
const Dotenv = require('dotenv-webpack');

const CopyPlugin = require('copy-webpack-plugin');
const pkg = require('../package.json');
const projectRoot = path.resolve(__dirname, '..');

/** @type {import('webpack').Configuration} */
module.exports = {
  entry: {
    main: './app',
  },
  output: {
    filename: '[name].[contenthash].js',
    path: path.resolve(projectRoot, 'dist/public'),
    pathinfo: false,
  },
  module: {
    rules: [
      {
        test: /\.js$/,
        type: 'javascript/auto',
        enforce: 'pre',
        exclude: /node_modules/,
        use: ['source-map-loader'],
      },
      {
        test: /\.(js|ts)(x)?$/,
        exclude: /node_modules/,
        use: ['babel-loader', 'auto-ngtemplate-loader'],
      },
      {
        test: /\.html$/,
        exclude: path.resolve(projectRoot, './app/index.html'),
        use: [
          {
            loader: 'ngtemplate-loader',
            options: {
              relativeTo: projectRoot + '/',
            },
          },
          { loader: 'html-loader' },
        ],
      },

      {
        test: /.xml$/,
        type: 'asset/resource',
      },
      {
        test: /\.(gif|png|jpe?g)$/i,
        type: 'asset/resource',
      },
      {
        test: /\.svg$/i,
        type: 'asset',
        resourceQuery: { not: [/c/] }, // exclude react component if *.svg?url
      },
      {
        test: /\.svg$/i,
        issuer: /\.(js|ts)(x)?$/,
        resourceQuery: /c/, // *.svg?c
        use: [{ loader: '@svgr/webpack', options: { icon: true } }],
      },
      {
        test: /\.css$/,
        use: [
          MiniCssExtractPlugin.loader,
          {
            loader: 'css-loader',
            options: {
              importLoaders: 1,
              modules: {
                localIdentName: '[path][name]__[local]',
                auto: true,
                exportLocalsConvention: 'camelCaseOnly',
              },
            },
          },
          {
            loader: 'postcss-loader',
          },
        ],
      },
      {
        test: /\.m?js/,
        resolve: {
          fullySpecified: false,
        },
      },
    ],
  },
  devServer: {
    static: {
      directory: path.join(__dirname, 'public'),
    },
    compress: true,
    port: 8999,
    proxy: {
      '/api': 'http://localhost:9000',
    },
    open: true,
    devMiddleware: {
      writeToDisk: true,
    },
  },
  plugins: [
    new Dotenv({ defaults: true }),
    new HtmlWebpackPlugin({
      template: './app/index.html',
      templateParameters: {
        name: pkg.name,
        author: pkg.author,
      },
      manifest: './assets/ico/manifest.json',
    }),
    new HtmlWebpackPlugin({
      template: './app/timeout.ejs',
      filename: 'timeout.html',
      templateParameters: {
        name: pkg.name,
        author: pkg.author,
      },
    }),
    new WebpackBuildNotifierPlugin({
      title: 'Portainer build',
      logo: path.resolve('./assets/favicon-32x32.png'),
      suppressSuccess: true,
    }),
    new ProvidePlugin({
      $: 'jquery',
      jQuery: 'jquery',
      'window.jQuery': 'jquery',
      'window.moment': 'moment',
      moment: 'moment',
      'window.jsyaml': 'js-yaml',
      jsyaml: 'js-yaml',
    }),
    new MiniCssExtractPlugin({
      filename: '[name].[contenthash].css',
      chunkFilename: '[name].[id].css',
    }),
    new CleanWebpackPlugin(),
    new IgnorePlugin({ resourceRegExp: /^.\/locale$/, contextRegExp: /moment$/ }),
    // new BundleAnalyzerPlugin()
    new LodashModuleReplacementPlugin({
      shorthands: true,
      collections: true,
      paths: true,
    }),
    new CopyPlugin({
      patterns: [
        {
          from: 'translations',
          to: 'locales',
        },
      ],
    }),
  ],
  optimization: {
    moduleIds: 'deterministic',
    runtimeChunk: 'single',
    splitChunks: {
      chunks: 'all',
      cacheGroups: {
        vendor: {
          test: /node_modules/,
          chunks: 'initial',
          name: 'vendor',
          priority: 10,
          enforce: true,
        },
      },
    },
  },
  watchOptions: {
    ignored: /node_modules/,
    aggregateTimeout: 200,
  },
  resolve: {
    alias: {
      '@@': path.resolve(projectRoot, 'app/react/components'),
      '@': path.resolve(projectRoot, 'app'),
      Agent: path.resolve(projectRoot, 'app/agent'),
      Azure: path.resolve(projectRoot, 'app/azure'),
      Docker: path.resolve(projectRoot, 'app/docker'),
      Kubernetes: path.resolve(projectRoot, 'app/kubernetes'),
      Portainer: path.resolve(projectRoot, 'app/portainer'),
      'lodash-es': 'lodash',
      'yaml-schema': path.resolve(projectRoot, 'node_modules/codemirror-json-schema/dist/yaml'),
    },
    extensions: ['.js', '.ts', '.tsx'],
    plugins: [
      new TsconfigPathsPlugin({
        extensions: ['.js', '.ts', '.tsx'],
      }),
    ],
  },
};
