/// 泛型快速排序实现
pub fn quick_sort<T: Ord>(arr: &mut [T]) {
    // 如果数组长度小于2，则已经有序
    if arr.len() <= 1 {
        return;
    }

    // 获取分区点索引
    let pivot_index = partition_hoare_unsafe(arr);

    // 递归排序左右两部分
    let (left, right) = arr.split_at_mut(pivot_index);
    quick_sort(left);
    quick_sort(&mut right[1..]); // 跳过基准元素
}

fn partition_lomuto<T: Ord>(arr: &mut [T]) -> usize {
    let len = arr.len();
    arr.swap(len/2, len - 1);

    // [0,i) all items smaller then arr[len-1]
    let mut i = 0;
    // move smaller items to front of arr
    for j in 0..len - 1 {
        // 将小于基准的元素移到左侧，如果本身是正序的，这里每次都要自己和自己swap，性能有问题
        if arr[j] <= arr[len - 1] {
            arr.swap(i, j);
            i += 1;
        }
    }

    arr.swap(i, len - 1);
    i
}

fn partition_hoare<T: Ord>(arr: &mut [T]) -> usize {
    let len = arr.len();
    arr.swap(0, len/2);  // 将基准移到开头

    let (mut left, mut right) = (1, len - 1);

    loop {
        // 从右向左找小于基准的元素
        while left <= right && arr[right] > arr[0] {
            right -= 1;
        }

        // 从左向右找大于基准的元素
        while left <= right && arr[left] < arr[0] {
            left += 1;
        }

        if left >= right {
            break;
        }

        // 交换并移动指针
        arr.swap(left, right);
        left += 1;
        right -= 1;
    }

    // 将基准放到正确位置
    arr.swap(0, right);
    right
}


fn partition_hoare_unsafe<T: Ord>(arr: &mut [T]) -> usize {
    unsafe{
        use std::ptr;
        let pivot = arr.as_mut_ptr();
        let mut left = pivot.add(1);
        let mut right = pivot.add(arr.len()-1);
        loop{
            while left <= right && *right > *pivot{
                right = right.sub(1);
            }
            while left <=right && *left < *pivot{
                left = left.add(1);
            }
            if left >= right {
                break;
            }
            ptr::swap(left, right);
            left = left.add(1);
            right = right.sub(1);
        }
        ptr::swap(pivot, right);
        right.offset_from(pivot) as usize
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_empty() {
        let mut arr: Vec<i32> = vec![];
        quick_sort(&mut arr);
        assert_eq!(arr, vec![]);
    }

    #[test]
    fn test_single_element() {
        let mut arr = vec![42];
        quick_sort(&mut arr);
        assert_eq!(arr, vec![42]);
    }

    #[test]
    fn test_sorted() {
        let mut arr = vec![1, 2, 3, 4, 5];
        quick_sort(&mut arr);
        assert_eq!(arr, vec![1, 2, 3, 4, 5]);
    }

    #[test]
    fn test_reverse_sorted() {
        let mut arr = vec![5, 4, 3, 2, 1];
        quick_sort(&mut arr);
        assert_eq!(arr, vec![1, 2, 3, 4, 5]);
    }

    #[test]
    fn test_random() {
        let mut arr = vec![3, 1, 4, 1, 5, 9, 2, 6];
        quick_sort(&mut arr);
        assert_eq!(arr, vec![1, 1, 2, 3, 4, 5, 6, 9]);
    }

    #[test]
    fn test_strings() {
        let mut arr = vec!["banana", "apple", "cherry", "date"];
        quick_sort(&mut arr);
        assert_eq!(arr, vec!["apple", "banana", "cherry", "date"]);
    }

    #[test]
    fn test_large_array() {
        let mut arr: Vec<i32> = (0..10000).rev().collect();
        quick_sort(&mut arr);
        let expected: Vec<i32> = (0..10000).collect();
        assert_eq!(arr, expected);
    }

    #[test]
    fn test_custom_struct() {
        #[derive(Debug, PartialEq, Eq, PartialOrd, Ord)]
        struct Person {
            name: String,
            age: u32,
        }

        let mut people = vec![
            Person { name: "Alice".to_string(), age: 30 },
            Person { name: "Bob".to_string(), age: 25 },
            Person { name: "Charlie".to_string(), age: 35 },
        ];

        quick_sort(&mut people);

        assert_eq!(people[0].name, "Alice");
        assert_eq!(people[1].name, "Bob");
        assert_eq!(people[2].name, "Charlie");
    }
}