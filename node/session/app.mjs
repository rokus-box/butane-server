'use strict';
import { DynamoDB } from '@aws-sdk/client-dynamodb';
import {  status } from './util.mjs';;

const TABLE_NAME = process.env.TABLE_NAME;
const dynamoDB = new DynamoDB({
  endpoint: 'http://192.168.1.12:8000', region: 'eu-central-1',
});

export const handler = async (event) => {
  const token = event.headers['authorization'];
  if (null == token) {
    return status(401);
  }

  const userId = await getUserId(dynamoDB, token);

  if ('' === userId) {
    return status(401);
  }

  return handleDeleteSession(userId, token);
};

async function getUserId(dynamoDB, token) {
  const sessionsRes = await dynamoDB.query({
    TableName: TABLE_NAME,
    KeyConditionExpression: 'PK = :pk',
    ProjectionExpression: 'SK',
    ExpressionAttributeValues: {
      ':pk': { S: 'SS#' + token },
    },
  }).catch((error) => {
    console.error('error getting session', error);
  });

  if (0 === sessionsRes.Items.length) {
    return '';
  }

  return sessionsRes.Items[0].SK.S;
}

async function handleDeleteSession(userId, token) {
  await dynamoDB.deleteItem({
    TableName: TABLE_NAME,
    Key: {
      PK: { S: 'SS#' + token },
      SK: { S: userId },
    },
  }).catch((error) => {
    console.error('error deleting session', error);
  });

  return status(204);
}
