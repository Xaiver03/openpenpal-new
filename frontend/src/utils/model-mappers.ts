// Utility functions for model field mapping

// Convert snake_case to camelCase
export function snakeToCamel<T extends Record<string, any>>(obj: T): any {
  if (Array.isArray(obj)) {
    return obj.map(item => snakeToCamel(item));
  }
  
  if (obj !== null && typeof obj === 'object') {
    return Object.keys(obj).reduce((result, key) => {
      const camelKey = key.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
      result[camelKey] = snakeToCamel(obj[key]);
      return result;
    }, {} as any);
  }
  
  return obj;
}

// Convert camelCase to snake_case
export function camelToSnake<T extends Record<string, any>>(obj: T): any {
  if (Array.isArray(obj)) {
    return obj.map(item => camelToSnake(item));
  }
  
  if (obj !== null && typeof obj === 'object') {
    return Object.keys(obj).reduce((result, key) => {
      const snakeKey = key.replace(/([A-Z])/g, '_$1').toLowerCase();
      result[snakeKey] = camelToSnake(obj[key]);
      return result;
    }, {} as any);
  }
  
  return obj;
}

// Type-safe model mappers
export const ModelMappers = {
  user: {
    fromAPI: (data: any): User => snakeToCamel(data),
    toAPI: (user: User): any => camelToSnake(user)
  },
  letter: {
    fromAPI: (data: any): Letter => snakeToCamel(data),
    toAPI: (letter: Letter): any => camelToSnake(letter)
  },
  courier: {
    fromAPI: (data: any): Courier => snakeToCamel(data),
    toAPI: (courier: Courier): any => camelToSnake(courier)
  }
};
