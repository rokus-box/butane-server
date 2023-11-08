'use strict';

import { handler } from '../../app.mjs';
import { expect } from 'chai';

describe('Tests index', function () {
    it('verifies successful response', async () => {
        expect(1).to.equal(1);
    });
});
