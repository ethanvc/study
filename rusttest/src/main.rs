
fn main() {
    example_ref_lib();
}

pub fn ffff(){
    let mut arr = vec![1, 2, 3];
    let ii = 3;
    unsafe{
        let pivot = arr.as_mut_ptr();
        let pivot_val = &*pivot;
        loop{
            if pivot_val > 3{
            }
        }
    }
}

fn example_ref_lib(){
    let mut vec = vec![1, 3, 2];
    rusttest::quick_sort::quick_sort( &mut vec);
    println!("exit")
}

fn example_change_vec_in_place(){
    let mut vals = vec![
        3, 4, 5,
    ];
    for val in &mut vals {
        *val *= 2;
    }
}

fn example_borrow_element_in_vec(){
    let vals = vec![
        String::from("hi"),
        String::from("from"),
        String::from("the"),
        String::from("future"),
    ];

    // ownership transferred to for block
    for val in vals {
        keep_element(val);
    }
    println!("exit");
}

fn keep_element(_s :String){

}

fn example_closure(){
    let mut count = 0;
    let mut increment = || {
        count += 1; // 需要可变引用
        println!("计数: {}", count);
    };
    increment(); // 输出: 计数: 1
    increment(); // 输出: 计数: 2
}
