'use strict';
import { DynamoDB } from '@aws-sdk/client-dynamodb';
import { json, status, text } from './util.mjs';
import { authenticator } from 'otplib';


const TABLE_NAME = process.env.TABLE_NAME;
const dynamoDB = new DynamoDB({
  endpoint: 'http://192.168.1.12:8000', region: 'eu-central-1',
});

export const handler = async (event) => {
  const token = event.headers['Authorization'];
  if (null == token) {
    return status(401);
  }

  const userId = await getUserId(event, dynamoDB, token);

  if ('' === userId) {
    return status(401);
  }

  switch (event.httpMethod) {
    case 'GET':
      return handleGetSessions(event, userId, token);
    case 'DELETE':
      const mfaCode = event.headers['X-Mfa-Challenge'];
      if (null == mfaCode) {
        return text('X-Mfa-Challenge header is required', 401);
      }

      const isValid = authenticator.verify({
        token: mfaCode,
        // TODO: Kullanicinin MFA kodunu DynamoDB'den al U#<userId>
        secret: 'JBSWY3DPEHPK3PXP',
      });

      console.log('isValid', isValid);

      return text('OK');
  }
};

async function handleGetSessions(event, userId, token) {
  const res = await dynamoDB.query({
    TableName: TABLE_NAME,
    KeyConditionExpression: 'PK = :pk AND SK = :sk',
    ProjectionExpression: 'ip_address, user_agent, timestamp',
    ExpressionAttributeValues: {
      ':pk': { S: 'SS#' + token },
      ':sk': { S: userId },
    },
  }).catch((error) => {
    console.error('error getting sessions', error);
  });

  return json(res.Items.map((item) => ({
    ip_address: item.ip_address.S,
    user_agent: item.user_agent.S,
    timestamp: Number(item.timestamp.N),
  })));
}

async function handleDeleteSessions(event, userId, token) {
  const res = await dynamoDB.query({
    TableName: TABLE_NAME,
    KeyConditionExpression: 'PK = :pk AND begins_with(SK, :sk)',
    ExpressionAttributeValues: {
      ':pk': { S: 'SS#' + token },
      ':sk': { S: userId },
    },
  }).catch((error) => {
    console.error('error getting sessions', error);
  });

  const promises = res.Items.map((item) => {
    return dynamoDB.deleteItem({
      TableName: TABLE_NAME,
      Key: {
        PK: { S: item.PK.S },
        SK: { S: item.SK.S },
      },
    });
  });

  await Promise.all(promises);

  return status(204);
}

async function getUserId(event, dynamoDB, token) {
  const res = await dynamoDB.query({
    TableName: TABLE_NAME,
    KeyConditionExpression: 'PK = :pk',
    ExpressionAttributeValues: {
      ':pk': { S: 'SS#' + token },
    },
  }).catch((error) => {
    console.error('error getting sessions', error);
  });

  if (0 === res.Items.length) {
    return '';
  }

  return res.Items[0].SK.S;
}
