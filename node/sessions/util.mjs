export function status(status) {
  return {
    statusCode: status,
    headers: {
      'Content-Type': 'text/plain',
    },
  };
}

export function text(text, status = 200) {
  return {
    statusCode: status,
    body: text,
    headers: {
      'Content-Type': 'text/plain',
    },
  };
}

export function json(data, status = 200) {
  return {
    statusCode: status,
    body: JSON.stringify(data),
    headers: {
      'Content-Type': 'application/json',
    },
  };
}
