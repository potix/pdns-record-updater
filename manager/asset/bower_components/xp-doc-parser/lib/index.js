/**
 * @license
 * Copyright (c) 2017 The expand.js authors. All rights reserved.
 * This code may only be used under the BSD style license found at https://expandjs.github.io/LICENSE.txt
 * The complete set of authors may be found at https://expandjs.github.io/AUTHORS.txt
 * The complete set of contributors may be found at https://expandjs.github.io/CONTRIBUTORS.txt
 */

// Const
const env     = typeof window !== "undefined" ? window : global,
    XP        = env.XP || require('expandjs'),
    XPEmitter = env.XPEmitter || require('xp-emitter'),
    XPRequest = env.XPRequest || require('xp-request');

/*********************************************************************/

/**
 * A class used to parse the documentation inside an html or javascript file.
 *
 * @class XPDocParser
 * @extends XPEmitter /bower_components/xp-emitter/lib/index.js
 * @since 1.0.0
 * @description A class used to parse the documentation inside an html or javascript file
 * @keywords nodejs, expandjs
 * @source https://github.com/expandjs/xp-doc-parser/blob/master/lib/index.js
 */
module.exports = env.XPDocParser = new XP.Class('XPDocParser', {

    // EXTENDS
    extends: XPEmitter,

    /*********************************************************************/

    /**
     * Emitted when data has been parsed.
     *
     * @event data
     * @param {Object} data
     */

    /**
     * Emitted when an error is received.
     *
     * @event error
     * @param {Object} error
     */

    /**
     * Emitted when the request's state changes.
     *
     * @event state
     * @param {string} state
     */

    /*********************************************************************/

    /**
     * @constructs
     * @param {Object | string} options The request's url or options
     *   @param {boolean} [options.inherited = false] Specifies if the data to parse is inherited
     *   @param {string} [options.url] The request's url
     * @param {Function} resolver The promise callback
     */
    initialize: {
        promise: true,
        value(options, resolver) {

            // Attempting
            XP.attempt(resolver => {

                // Super
                XPEmitter.call(this);

                // Overriding
                if (!XP.isObject(options)) { options = {url: options}; }

                // Setting
                this.options   = options;
                this.url       = this.options.url;
                this.request   = new XPRequest({parser: 'text', url: this.url});
                this.state     = this.request.state;

                // Listening
                this.request.on('data', this._handleData.bind(this));
                this.request.on('error', error => this.error = error);
                this.request.on('state', state => this.state = state);

                // Resolving
                this.on('data', data => resolver(null, data));
                this.on('error', error => resolver(error, null));

                // Submitting
                this.request.submit();

            }, resolver);
        }
    },

    /*********************************************************************/

    /**
     * Parses text.
     *
     * @method parse
     * @param {string} [text = ""]
     * @returns {Object}
     */
    parse(text) {

        // Asserting
        XP.assertArgument(XP.isVoid(text) || XP.isString(text), 1, 'string');

        // Let
        let regExpScript = '\\/\\*\\*([\\s\\S]*?)\\*\\/',
            regExpHtml   = '<!--([\\s\\S]*?)-->',
            regExp       = text ? new RegExp(`${regExpScript}|${regExpHtml}`, `g`) : null,
            blocks       = text ? XP.match(text, regExp).map(block => this.parseBlock(block)) : [],
            entity       = text ? blocks.find(block => this.entities.indexOf(block.type) >= 0) : null;

        // Interpreting
        if (entity) { this._interpretEntity(entity, XP.pull(blocks, entity)); }

        // Returning
        return entity || {type: 'Unknown', name: 'Unknown', summary: '**Undocumented**'};
    },

    /**
     * Parses a comment block.
     *
     * @method parseBlock
     * @param {string} [block = ""]
     * @returns {Object}
     */
    parseBlock(block) {

        // Asserting
        XP.assertArgument(XP.isVoid(block) || XP.isString(block), 1, 'string');

        // Preparing
        block = block || '';
        block = block.replace(/\r\n/g, '\n'); // Unifies line ends
        block = block.replace(/^\s*\/\*\*|^\s*\*\/|^\s*\* ?|^\s*<!-\-|^s*\-\->/gm, ''); // Removes all comment characters

        // Let
        let lines  = block.split('\n'),
            result = {};

        // Filtering
        lines = lines.filter(line => {

            // Let
            let parts = XP.match(line, /^@(\w+)(.*)|^\s+\s+@(\w+)(.*)/);

            // Checking
            if (!parts.length) { return true; }

            // Pushing
            if (result.pragma) {
                result.pragma.push({type: parts[1] || parts[3], value: XP.clean(parts[2] || parts[4]) || true});
            }

            // Setting
            if (!result.pragma) {
                result.type   = parts[1] || parts[3];
                result.name   = XP.clean(parts[2] || parts[4]);
                result.pragma = [];
            }
        });

        // Setting
        result.type    = result.type || '';
        result.name    = result.name || '';
        result.summary = lines.join('\n').trim();

        // Returning
        return result;
    },

    /*********************************************************************/

    /**
     * Returns the right interpreter based on the provided `name`.
     *
     * @method _computeInterpreter
     * @param {string} name
     * @returns {*}
     * @private
     */
    _computeInterpreter(name) {

        // Let
        let result = `_interpret${XP.capitalize(name)}`;

        // Returning
        return XP.isFunction(this[result]) ? result : undefined;
    },

    /**
     * Interprets an `adapts` pragma.
     *
     * @method _interpretAdapts
     * @param {Object} context
     * @param {Object} pragma
     * @private
     */
    _interpretAdapts(context, pragma) {

        // Super
        this._interpretDependency(context, pragma);
    },

    /**
     * Interprets a `behavior` pragma.
     *
     * @method _interpretBehavior
     * @param {Object} context
     * @param {Object} pragma
     * @private
     */
    _interpretBehavior(context, pragma) {

        // Super
        this._interpretDependency(context, pragma);
    },

    /**
     * Interprets a `dependency` pragma.
     *
     * @method _interpretDependency
     * @param {Object} context
     * @param {Object} pragma
     * @private
     */
    _interpretDependency(context, pragma) {

        // Let
        let parts = pragma.value.split(' ');

        // Overriding
        pragma.value = {
            name: XP.trim(parts[0]),
            url: XP.trim(parts[1] || '')
        };

        // Super
        this._interpretPragma(context, pragma);
    },

    /**
     * Interprets a `devDependency` pragma.
     *
     * @method _interpretDevDependency
     * @param {Object} context
     * @param {Object} pragma
     * @private
     */
    _interpretDevDependency(context, pragma) {

        // Super
        this._interpretDependency(context, pragma);
    },

    /**
     * Interprets an entity.
     *
     * @method _interpretEntity
     * @param {Object} entity
     * @param {Array} [features]
     * @private
     */
    _interpretEntity(entity, features) {

        // Asserting
        XP.assertArgument(XP.isObject(entity), 1, 'Object');
        XP.assertArgument(XP.isVoid(features) || XP.isArray(features), 2, 'ArrayLike');

        // Let
        let pragma = XP.withdraw(entity, 'pragma');

        // Interpreting
        if (pragma) { pragma.forEach(pragma => XP.apply(this, this._computeInterpreter(pragma.type) || '_interpretPragma', [entity, pragma])); }
        if (features) { features.forEach(feature => XP.apply(this, this._computeInterpreter(feature.type) || '_interpretFeature', [entity, feature])); }
    },

    /**
     * Interprets an `extends` pragma.
     *
     * @method _interpretExtends
     * @param {Object} context
     * @param {Object} pragma
     * @private
     */
    _interpretExtends(context, pragma) {

        // Super
        this._interpretDependency(context, pragma);
    },

    /**
     * Interprets a feature.
     *
     * @method _interpretFeature
     * @param {Object} entity
     * @param {Object} feature
     * @private
     */
    _interpretFeature(entity, feature) {

        // Asserting
        XP.assertArgument(XP.isObject(entity), 1, 'Object');
        XP.assertArgument(XP.isObject(feature), 2, 'Object');

        // Let
        let name   = feature.name || XP.withdraw(feature, 'name'),
            pragma = XP.withdraw(feature, 'pragma'),
            type   = XP.withdraw(feature, 'type'),
            plural = this.plurals[type || ''];

        // Checking
        if (!type) { return; }

        // Interpreting
        if (pragma) { pragma.forEach(pragma => XP.apply(this, this._computeInterpreter(pragma.type) || '_interpretPragma', [feature, pragma])); }

        // Setting
        if (plural) { (entity[plural] = entity[plural] || []).push(feature); } else { entity[type] = feature; }
    },

    /**
     * Interprets a `keywords` pragma.
     *
     * @method _interpretKeywords
     * @param {Object} context
     * @param {Object} pragma
     * @private
     */
    _interpretKeywords(context, pragma) {

        // Overriding
        pragma.value = XP.split(pragma.value, ',').map(keyword => keyword.trim());

        // Super
        this._interpretPragma(context, pragma);
    },

    /**
     * Interprets a `param` pragma.
     *
     * @method _interpretParam
     * @param {Object} context
     * @param {Object} pragma
     * @private
     */
    _interpretParam(context, pragma) {

        // Let
        let regExp = /\{(.+)\}\s+(\[.+\]|\w+\.\w+|\w+)\s*(.*)$/,
            parts1 = XP.match(pragma.value, regExp),
            parts2 = XP.split(XP.trim(parts1[2], '[]'), '=', true),
            parts3 = XP.split(XP.trim(parts2[0]), '.', true);

        // Overriding
        pragma.value = {
            type: XP.trim(parts1[1]),
            name: XP.trim(parts3[parts3.length > 1 ? 1 : 0]),
            summary: XP.trim(parts1[3]),
            optional: XP.startsWith(parts1[2], '['),
            default: XP.trim(parts2[1])
        };

        // Deleting
        if (!pragma.value.optional) { delete pragma.value.optional; }
        if (!pragma.value.default) { delete pragma.value.default; }

        // Context
        if (parts3.length > 1) { context = XP.findLast(context.params || [], param => param.name === parts3[0].trim()) || {}; }

        // Super
        this._interpretPragma(context, pragma);
    },

    /**
     * Interprets a pragma.
     *
     * @method _interpretPragma
     * @param {Object} context
     * @param {Object} pragma
     * @private
     */
    _interpretPragma(context, pragma) {

        // Asserting
        XP.assertArgument(XP.isObject(context), 1, 'Object');
        XP.assertArgument(XP.isObject(pragma), 2, 'Object');

        // Let
        let type   = XP.withdraw(pragma, 'type'),
            plural = this.plurals[type || ''];

        // Setting
        if (plural) {
            context[plural] = context[plural] || [];
            context[plural].push(pragma.value);
        } else {
            context[type] = pragma.value;
        }
    },

    /**
     * Interprets a `returns` pragma.
     *
     * @method _interpretReturns
     * @param {Object} context
     * @param {Object} pragma
     * @private
     */
    _interpretReturns(context, pragma) {

        // Let
        let regExp = /\{(.+)\}\s*(.*)$/,
            parts  = pragma.value.match(regExp) || [];

        // Overriding
        pragma.value = {
            type: XP.trim(parts[1]),
            summary: XP.trim(parts[2])
        };

        // Super
        this._interpretPragma(context, pragma);
    },

    /**
     * Merges the provided `data` with `parsers` data.
     *
     * @method _mergeData
     * @param {Object} data
     * @param {Array} parsers
     * @returns {Object}
     * @private
     */
    _mergeData(data, parsers) {

        // Asserting
        XP.assertArgument(XP.isObject(data), 1, 'Object');
        XP.assertArgument(XP.isArray(parsers), 2, 'Array');

        // Merging
        parsers.reverse().forEach(parser => {

            // Inheritance
            this.inheritance.forEach(feature => parser.data[feature] && parser.data[feature].forEach(feat => {
                if (Array.from(data[feature] || []).some(item => item.name === feat.name)) { return; }
                data[feature] = data[feature] || [];
                data[feature].push(Object.assign(feat, {inherited: true}));
            }));
        });

        // Returning
        return data;
    },

    /**
     * Sorts the provided `data`.
     *
     * @method _sortData
     * @param {Object} data
     * @param {boolean} [useful = false]
     * @returns {Object}
     * @private
     */
    _sortData(data, useful) {

        // Asserting
        XP.assertArgument(XP.isObject(data), 1, 'Object');

        // Checking
        if (useful && this.inherited) { return data; }

        // Sorting
        this.inheritance.forEach(feature => {
            return data[feature] && data[feature].sort((a, b) => {
                if (!a.private && b.private) { return 1; }
                if (a.private && !b.private) { return -1; }
                if (a.name > b.name) { return 1; }
                if (a.name < b.name) { return -1; }
            });
        });

        // Returning
        return data;
    },

    /*********************************************************************/

    /**
     * The parsed data.
     *
     * @property data
     * @type Object
     * @readonly
     */
    data: {
        set(val) { return XP.isDefined(this.data) ? this.data : val; },
        then(val) { return val && this.emit('data', val); },
        validate(val) { return !XP.isNull(val) && !XP.isObject(val) && 'Object'; }
    },

    /**
     * If set to true, the data are not ready yet.
     *
     * @property empty
     * @type boolean
     * @default true
     * @readonly
     */
    empty: {
        get() { return !this.data; }
    },

    /**
     * The list of possible entities.
     *
     * @property entities
     * @type Array
     * @default ["behavior", "class", "element", "function", "module", "object", "stylesheet"]
     * @readonly
     */
    entities: {
        frozen: true,
        writable: false,
        value: ['behavior', 'class', 'element', 'function', 'module', 'object', 'stylesheet']
    },

    /**
     * The request's error.
     *
     * @property error
     * @type Object
     * @readonly
     */
    error: {
        set(val) { return XP.isDefined(this.error) ? this.error : val; },
        then(val) { return val && this.emit('error', val); },
        validate(val) { return !XP.isNull(val) && !XP.isObject(val) && 'Object'; }
    },

    /**
     * The request's host.
     *
     * @property host
     * @type string
     * @readonly
     */
    host: {
        set(val) { return this.host || val; },
        validate(val) { return !XP.isString(val, true) && 'string'; }
    },

    /**
     * The list of possible inheritance.
     *
     * @property inheritance
     * @type Array
     * @default ["attributes", "events", "methods", "properties"]
     * @readonly
     */
    inheritance: {
        frozen: true,
        writable: false,
        value: ['attributes', 'events', 'methods', 'properties']
    },

    /**
     * If set to true, the data to parse is inherited.
     *
     * @property inherited
     * @type boolean
     * @default false
     * @readonly
     */
    inherited: {
        set(val) { return XP.isDefined(this.inherited) ? this.inherited : !!val; }
    },

    /**
     * The map of plurals.
     *
     * @property plurals
     * @type Object
     * @readonly
     */
    plurals: {
        frozen: true,
        writable: false,
        value: {
            attribute: 'attributes',
            behavior: 'behaviors',
            dependency: 'dependencies',
            devDependency: 'devDependencies',
            event: 'events',
            method: 'methods',
            param: 'params',
            property: 'properties'
        }
    },

    /**
     * The request's protocol.
     *
     * @property protocol
     * @type string
     * @readonly
     */
    protocol: {
        set(val) { return this.protocol || val; },
        validate(val) { return !XP.isString(val, true) && 'string'; }
    },

    /**
     * The request's state.
     *
     * @property state
     * @type string
     * @default "idle"
     * @readonly
     */
    state: {
        set(val) { return val; },
        then(val) { return val !== 'idle' && this.emit('state', val); },
        validate(val) { return !XP.isString(val, true) && 'string'; }
    },

    /**
     * The documentation's URL.
     *
     * @property url
     * @type string
     */
    url: {
        set(val) { return this.url || val; },
        then(val) { let parsed = XP.parseURL(val); this.host = parsed.host; this.protocol = parsed.protocol; },
        validate(val) { return !XP.isString(val, true) && 'string'; }
    },

    /*********************************************************************/

    // HANDLER
    _handleData(data) {

        // Let
        let parsed  = data ? this.parse(data) : null,
            parser  = null,
            parsers = [],
            count   = 0;

        // Merging (extends)
        if (parsed && parsed.extends && parsed.extends.url) {
            let object = XP.parseURL(parsed.extends.url);
            object.host = object.host || this.host;
            object.protocol = object.protocol || this.protocol;
            parser = new module.exports({inherited: true, url: XP.toURL(object)});
            parser.on('data', () => (count += 1) === parsers.length && (this.data = this._sortData(this._mergeData(parsed, parsers), true)));
            parsers.push(parser);
        }

        // Merging (behaviors)
        if (parsed && parsed.behaviors) {
            parsed.behaviors.forEach(behavior => {
                if (!behavior.url) { return; }
                let object = XP.parseURL(behavior.url);
                object.host = object.host || this.host;
                object.protocol = object.protocol || this.protocol;
                parser = new module.exports({inherited: true, url: XP.toURL(object)});
                parser.on('data', () => (count += 1) === parsers.length && (this.data = this._sortData(this._mergeData(parsed, parsers), true)));
                parsers.push(parser);
            });
        }

        // Setting
        if (!parsers.length) { this.data = parsed && this._sortData(parsed, true); }
    }
});
